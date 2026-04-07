package model

type Severity string

const (
	LOW    Severity = "LOW"
	MEDIUM Severity = "MEDIUM"
	HIGH   Severity = "HIGH"
)

var SeverityRank = map[Severity]int{
	LOW:    1,
	MEDIUM: 2,
	HIGH:   3,
}

type Issue struct {
	Severity       Severity `json:"severity"`
	Message        string   `json:"message"`
	Recommendation string   `json:"recommendation"`
	Path           string   `json:"path"`
}
