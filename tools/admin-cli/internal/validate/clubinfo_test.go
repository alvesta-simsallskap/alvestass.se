package validate

import (
	"testing"

	"github.com/alvestass/admin-cli/internal/trailbase"
	"github.com/stretchr/testify/assert"
)

func validRecord() *trailbase.ClubInfo {
	return &trailbase.ClubInfo{
		ID:               1,
		Name:             "Alvesta Simsällskap",
		Tagline:          "Simglädje sedan 1921",
		FoundingYear:     1921,
		ShortDescription: "Simglädje.",
		Address:          "Hjortsbergavägen 6C",
		City:             "Alvesta",
		PostalCode:       "342 36",
		Phone:            "076 027 94 10",
		Email:            "kansli@alvestass.se",
	}
}

func TestValidRecordNoIssues(t *testing.T) {
	issues := ValidateClubInfo(validRecord())
	assert.Empty(t, issues)
}

func TestEmptyNameReturnsIssue(t *testing.T) {
	r := validRecord()
	r.Name = ""
	issues := ValidateClubInfo(r)
	assertField(t, issues, "name")
}

func TestEmptyAddressReturnsIssue(t *testing.T) {
	r := validRecord()
	r.Address = "   "
	issues := ValidateClubInfo(r)
	assertField(t, issues, "address")
}

func TestEmptyCityReturnsIssue(t *testing.T) {
	r := validRecord()
	r.City = ""
	issues := ValidateClubInfo(r)
	assertField(t, issues, "city")
}

func TestEmptyPostalCodeReturnsIssue(t *testing.T) {
	r := validRecord()
	r.PostalCode = ""
	issues := ValidateClubInfo(r)
	assertField(t, issues, "postal_code")
}

func TestEmptyPhoneReturnsIssue(t *testing.T) {
	r := validRecord()
	r.Phone = ""
	issues := ValidateClubInfo(r)
	assertField(t, issues, "phone")
}

func TestEmptyEmailReturnsIssue(t *testing.T) {
	r := validRecord()
	r.Email = ""
	issues := ValidateClubInfo(r)
	assertField(t, issues, "email")
}

func TestInvalidEmailFormatReturnsIssue(t *testing.T) {
	r := validRecord()
	r.Email = "not-an-email"
	issues := ValidateClubInfo(r)
	assertField(t, issues, "email")
}

func TestValidEmailFormats(t *testing.T) {
	for _, email := range []string{"a@b.se", "user+tag@example.com", "kansli@alvestass.se"} {
		r := validRecord()
		r.Email = email
		assert.Empty(t, ValidateClubInfo(r), "expected valid email %q to pass", email)
	}
}

func TestFoundingYearBoundaries(t *testing.T) {
	r := validRecord()

	r.FoundingYear = 1800
	assert.Empty(t, ValidateClubInfo(r))

	r.FoundingYear = 2100
	assert.Empty(t, ValidateClubInfo(r))

	r.FoundingYear = 1799
	assertField(t, ValidateClubInfo(r), "founding_year")

	r.FoundingYear = 2101
	assertField(t, ValidateClubInfo(r), "founding_year")
}

func TestShortDescriptionMaxLength(t *testing.T) {
	r := validRecord()

	r.ShortDescription = string(make([]rune, 300))
	assert.Empty(t, ValidateClubInfo(r))

	r.ShortDescription = string(make([]rune, 301))
	assertField(t, ValidateClubInfo(r), "short_description")
}

func TestPostalCodeFormats(t *testing.T) {
	r := validRecord()

	for _, valid := range []string{"342 36", "34236", "123 45"} {
		r.PostalCode = valid
		assert.Empty(t, ValidateClubInfo(r), "expected valid postal code %q to pass", valid)
	}

	for _, invalid := range []string{"3423", "abc de", "12345 6"} {
		r.PostalCode = invalid
		assertField(t, ValidateClubInfo(r), "postal_code")
	}
}

func TestMultipleViolationsAllReported(t *testing.T) {
	r := &trailbase.ClubInfo{ID: 1, FoundingYear: 1921}
	issues := ValidateClubInfo(r)
	assert.GreaterOrEqual(t, len(issues), 4, "name, address, city, phone, email should all fail")
}

// assertField checks that at least one issue targets the given field.
func assertField(t *testing.T, issues []CheckIssue, field string) {
	t.Helper()
	for _, iss := range issues {
		if iss.Field == field {
			return
		}
	}
	t.Errorf("expected an issue for field %q, got %v", field, issues)
}
