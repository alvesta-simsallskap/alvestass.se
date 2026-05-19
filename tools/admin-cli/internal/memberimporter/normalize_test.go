package memberimporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeGroupName_StripsTimeSlotDot(t *testing.T) {
	assert.Equal(t, "Baddaren", NormalizeGroupName("Baddaren 12.55-13.40"))
}

func TestNormalizeGroupName_StripsTimeSlotDash(t *testing.T) {
	assert.Equal(t, "Guldpingvinen", NormalizeGroupName("Guldpingvinen 13.05–13.50"))
}

func TestNormalizeGroupName_StripsTimeSlotColon(t *testing.T) {
	assert.Equal(t, "Silverfisken", NormalizeGroupName("Silverfisken 08:00-08:45"))
}

func TestNormalizeGroupName_StripsTimeSlotSingleDigitHour(t *testing.T) {
	assert.Equal(t, "Guldfisken", NormalizeGroupName("Guldfisken 9.00-9.45"))
}

func TestNormalizeGroupName_NoOpWithoutTimeSlot(t *testing.T) {
	assert.Equal(t, "Masters", NormalizeGroupName("Masters"))
	assert.Equal(t, "A-gruppen", NormalizeGroupName("A-gruppen"))
	assert.Equal(t, "Teknikgruppen", NormalizeGroupName("Teknikgruppen"))
}

func TestNormalizeGroupName_TrimsWhitespace(t *testing.T) {
	assert.Equal(t, "Baddaren", NormalizeGroupName("  Baddaren 10.00-10.45  "))
}

var categoryTests = []struct {
	name     string
	expected string
}{
	{"Baddaren", "swim_school"},
	{"Guldfisken", "swim_school"},
	{"Guldhajen", "swim_school"},
	{"Guldpingvinen", "swim_school"},
	{"Silverfisken", "swim_school"},
	{"Silverhajen", "swim_school"},
	{"Silverpingvinen", "swim_school"},
	{"A-gruppen", "competitive"},
	{"B-gruppen", "competitive"},
	{"Teknikgruppen", "technique"},
	{"Masters", "masters"},
	{"Vuxencrawl", "adult"},
}

func TestGroupCategory_AllMappings(t *testing.T) {
	for _, tc := range categoryTests {
		t.Run(tc.name, func(t *testing.T) {
			got := GroupCategory(tc.name)
			assert.Equal(t, tc.expected, got, "GroupCategory(%q)", tc.name)
		})
	}
}

func TestGroupCategory_UnknownReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", GroupCategory("OkändGrupp"))
}
