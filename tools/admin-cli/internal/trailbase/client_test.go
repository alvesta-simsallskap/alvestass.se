package trailbase

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClubInfoMarshal(t *testing.T) {
	ci := ClubInfo{
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

	data, err := json.Marshal(ci)
	require.NoError(t, err)

	var got ClubInfo
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, ci, got)
}

func TestClubInfoJSONFieldNames(t *testing.T) {
	ci := ClubInfo{Name: "Test", FoundingYear: 1921, PostalCode: "123 45"}
	data, err := json.Marshal(ci)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))

	assert.Contains(t, m, "founding_year", "JSON key must use snake_case to match Trailbase schema")
	assert.Contains(t, m, "postal_code")
	assert.Contains(t, m, "short_description")
}
