package memberimporter

import (
	"regexp"
	"strings"
)

// timeSlotRegex matches a time-slot suffix such as " 12.55-13.40" or " 13.05–13.50".
var timeSlotRegex = regexp.MustCompile(`\s+\d{1,2}[.:]\d{2}\s*[-–]\s*\d{1,2}[.:]\d{2}$`)

// NormalizeGroupName strips any trailing time-slot suffix from a group name.
// "Baddaren 12.55-13.40" → "Baddaren"
func NormalizeGroupName(raw string) string {
	return strings.TrimSpace(timeSlotRegex.ReplaceAllString(strings.TrimSpace(raw), ""))
}

// groupCategoryMap maps normalised group names to Trailbase category values.
var groupCategoryMap = map[string]string{
	"Baddaren":        "swim_school",
	"Guldfisken":      "swim_school",
	"Guldhajen":       "swim_school",
	"Guldpingvinen":   "swim_school",
	"Silverfisken":    "swim_school",
	"Silverhajen":     "swim_school",
	"Silverpingvinen": "swim_school",
	"A-gruppen":       "competitive",
	"B-gruppen":       "competitive",
	"Teknikgruppen":   "technique",
	"Masters":         "masters",
	"Vuxencrawl":      "adult",
}

// GroupCategory returns the Trailbase category value for a normalised group name.
// Returns "" for unknown names.
func GroupCategory(name string) string {
	return groupCategoryMap[name]
}
