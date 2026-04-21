package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const installScriptURL = "https://raw.githubusercontent.com/richardriman/crafter/main/install.sh"

var updateVersion string
var updateLocal bool
var updateGlobal bool

var updateVersionPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Crafter installation using the official installer",
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "specific version to install (e.g. 0.8.1)")
	updateCmd.Flags().BoolVar(&updateLocal, "local", false, "update local project installation (.claude/)")
	updateCmd.Flags().BoolVar(&updateGlobal, "global", false, "update global installation (~/.claude/) (default)")

	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateLocal && updateGlobal {
		return fmt.Errorf("choose only one mode: --local or --global")
	}

	mode := "--global"
	if updateLocal {
		mode = "--local"
	}

	version := strings.TrimSpace(updateVersion)
	version = strings.TrimPrefix(version, "v")
	if version != "" && !updateVersionPattern.MatchString(version) {
		return fmt.Errorf("invalid version format: %q", updateVersion)
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		return fmt.Errorf("bash not found in PATH: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "crafter-install-*.sh")
	if err != nil {
		return fmt.Errorf("creating temp installer file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Get(installScriptURL)
	if err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("downloading installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_ = tmpFile.Close()
		return fmt.Errorf("downloading installer: unexpected HTTP status %s", resp.Status)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing installer file: %w", err)
	}
	if err := tmpFile.Chmod(0755); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("making installer executable: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing installer file: %w", err)
	}

	installerArgs := []string{tmpPath, mode}
	if version != "" {
		installerArgs = append(installerArgs, "--version", version)
	}

	fmt.Fprintln(os.Stderr, "Updating Crafter...")
	update := exec.Command(bashPath, installerArgs...)
	update.Stdout = os.Stdout
	update.Stderr = os.Stderr
	update.Stdin = os.Stdin

	if err := update.Run(); err != nil {
		return fmt.Errorf("installer failed: %w", err)
	}

	return nil
}
