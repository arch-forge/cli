package domain

// Severity classifies how serious a violation is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Violation describes a single architecture rule breach.
type Violation struct {
	File     string
	Line     int
	Rule     string
	Message  string
	Severity Severity
}

// Report is the output of a full doctor analysis run.
type Report struct {
	ProjectPath string
	Arch        Architecture
	Variant     Variant
	Violations  []Violation
	TotalRules  int
	Score       float64 // computed by ComputeScore()
}

// ComputeScore calculates score as (TotalRules - errorCount) / TotalRules * 10.
// Returns 10.0 if TotalRules == 0.
func (r *Report) ComputeScore() {
	if r.TotalRules == 0 {
		r.Score = 10.0
		return
	}
	errorCount := 0
	for _, v := range r.Violations {
		if v.Severity == SeverityError {
			errorCount++
		}
	}
	r.Score = float64(r.TotalRules-errorCount) / float64(r.TotalRules) * 10.0
}

// HasErrors reports whether any violation has SeverityError.
func (r Report) HasErrors() bool {
	for _, v := range r.Violations {
		if v.Severity == SeverityError {
			return true
		}
	}
	return false
}
