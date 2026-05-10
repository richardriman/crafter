package prbody

// UATEntry is the subset of the buffer UAT schema used for PR-body rendering.
// Fields not needed for rendering (ID, Kind, CreatedAt, CreatedBy, TaskID, Source)
// are decoded but ignored; Source is explicitly excluded from the rendered output
// per Phase 1 Decision 4(C).
type UATEntry struct {
	Title     string `json:"title"`
	Verify    string `json:"verify"`
	WhyManual string `json:"why_manual"`
}

// GapEntry is the subset of the buffer Gap schema used for PR-body rendering.
// Source is explicitly excluded from the rendered output per Phase 1 Decision 4(C).
type GapEntry struct {
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Followup string `json:"followup"`
}

// markerLine is the JSON structure used to detect a buffer marker line.
// Only the _marker field is inspected; other fields are ignored.
type markerLine struct {
	Marker string `json:"_marker"`
}
