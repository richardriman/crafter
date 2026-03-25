package skillbook

// Skillbook is the top-level structure stored in skillbook.json.
type Skillbook struct {
	Version   int     `json:"version"`
	UpdatedAt string  `json:"updatedAt"`
	Disabled  bool    `json:"disabled,omitempty"`
	Skills    []Skill `json:"skills"`
}

// Skill represents a single learned skill entry.
type Skill struct {
	ID           string `json:"id"`
	Agent        string `json:"agent"`
	Rule         string `json:"rule"`
	Rationale    string `json:"rationale"`
	SourceTask   string `json:"sourceTask"`
	Confidence   string `json:"confidence"`   // "low", "medium", "high"
	AddedAt      string `json:"addedAt"`
	AppliedCount int    `json:"appliedCount"`
	Deprecated   bool   `json:"deprecated"`
}
