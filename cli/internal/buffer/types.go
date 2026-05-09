package buffer

// UATEntry represents a single UAT (manual-QA) buffer entry.
// All fields are required in schema-1; none use omitempty.
type UATEntry struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
	TaskID    string `json:"task_id"`
	Title     string `json:"title"`
	Source    string `json:"source"`
	Verify    string `json:"verify"`
	WhyManual string `json:"why_manual"`
}

// GapEntry represents a single Gap (tech-debt / deferred follow-up) buffer entry.
// All fields are required in schema-1; none use omitempty.
type GapEntry struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
	TaskID    string `json:"task_id"`
	Title     string `json:"title"`
	Source    string `json:"source"`
	Detail    string `json:"detail"`
	Followup  string `json:"followup"`
}
