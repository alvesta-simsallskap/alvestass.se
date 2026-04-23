package validate

import "context"

// CheckIssue describes a single data quality problem found by a Checker.
type CheckIssue struct {
	Entity string // e.g. "club_info/1", "member/42"
	Field  string // e.g. "email"
	Value  string // the offending value (redact PII for future personal-data checkers)
	Rule   string // Swedish human-readable rule that was violated
}

// Checker is the extensibility contract for health checks.
// Every present and future check implements this interface.
// The runner in ui/check.go iterates a []Checker slice and groups results by Name().
type Checker interface {
	// Name returns the display name shown in the grouped results report.
	Name() string
	// Run fetches the relevant data and returns any issues found.
	// A non-nil error means the data could not be fetched; the runner
	// will report the error inline and continue with remaining checkers.
	Run(ctx context.Context) ([]CheckIssue, error)
}
