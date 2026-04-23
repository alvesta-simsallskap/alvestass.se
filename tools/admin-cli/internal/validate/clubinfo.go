package validate

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/alvestass/admin-cli/internal/trailbase"
)

var (
	emailRe      = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	postalCodeRe = regexp.MustCompile(`^\d{3}\s?\d{2}$`)
)

// ClubInfoChecker validates the single club_info record.
type ClubInfoChecker struct {
	client *trailbase.Client
}

// NewClubInfoChecker returns a ClubInfoChecker using the given Trailbase client.
func NewClubInfoChecker(client *trailbase.Client) *ClubInfoChecker {
	return &ClubInfoChecker{client: client}
}

func (c *ClubInfoChecker) Name() string { return "Kontaktuppgifter" }

func (c *ClubInfoChecker) Run(_ context.Context) ([]CheckIssue, error) {
	info, err := c.client.GetClubInfo()
	if err != nil {
		return nil, err
	}
	return ValidateClubInfo(info), nil
}

// ValidateClubInfo runs all field-level rules against a ClubInfo record and
// returns a CheckIssue for each violation. A nil/empty slice means no issues.
func ValidateClubInfo(info *trailbase.ClubInfo) []CheckIssue {
	var issues []CheckIssue

	entity := fmt.Sprintf("club_info/%d", info.ID)

	requireNonEmpty := func(field, value, label string) {
		if strings.TrimSpace(value) == "" {
			issues = append(issues, CheckIssue{
				Entity: entity,
				Field:  field,
				Value:  value,
				Rule:   label + " får inte vara tomt",
			})
		}
	}

	requireNonEmpty("name", info.Name, "Namn")
	requireNonEmpty("address", info.Address, "Adress")
	requireNonEmpty("city", info.City, "Stad")
	requireNonEmpty("postal_code", info.PostalCode, "Postnummer")
	requireNonEmpty("phone", info.Phone, "Telefonnummer")
	requireNonEmpty("email", info.Email, "E-postadress")

	if info.FoundingYear < 1800 || info.FoundingYear > 2100 {
		issues = append(issues, CheckIssue{
			Entity: entity,
			Field:  "founding_year",
			Value:  fmt.Sprintf("%d", info.FoundingYear),
			Rule:   "Grundårtal måste vara mellan 1800 och 2100",
		})
	}

	if len([]rune(info.ShortDescription)) > 300 {
		issues = append(issues, CheckIssue{
			Entity: entity,
			Field:  "short_description",
			Value:  fmt.Sprintf("(%d tecken)", len([]rune(info.ShortDescription))),
			Rule:   "Kort beskrivning får inte vara längre än 300 tecken",
		})
	}

	if strings.TrimSpace(info.Email) != "" && !emailRe.MatchString(info.Email) {
		issues = append(issues, CheckIssue{
			Entity: entity,
			Field:  "email",
			Value:  info.Email,
			Rule:   "E-postadress har ogiltigt format",
		})
	}

	if strings.TrimSpace(info.PostalCode) != "" && !postalCodeRe.MatchString(info.PostalCode) {
		issues = append(issues, CheckIssue{
			Entity: entity,
			Field:  "postal_code",
			Value:  info.PostalCode,
			Rule:   "Postnummer måste ha formatet NNN NN",
		})
	}

	return issues
}
