package memberimporter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wuHeader() string {
	return "Personnummer;Roll;Grupp;Start;Slut;" +
		"Målsman 1, Förnamn;Målsman 1, Efternamn;Målsman 1, Telefon;Målsman 1, E-post;" +
		"Målsman 2, Förnamn;Målsman 2, Efternamn;Målsman 2, Telefon;Målsman 2, E-post;" +
		"Målsman 3, Förnamn;Målsman 3, Efternamn;Målsman 3, Telefon;Målsman 3, E-post"
}

func TestParseWeUnite_DeltagareCollected(t *testing.T) {
	csv := wuHeader() + "\n" +
		"20100315-1234;Deltagare;Baddaren 12.55-13.40;2025-09-01;2026-06-01;Mamma;Svensson;070-1111111;mamma@example.com;;;;;;;;;\n"

	deltagare, instructors, guardians, _, err := ParseWeUniteReader(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, deltagare, 1)
	assert.Empty(t, instructors)

	d := deltagare[0]
	assert.Equal(t, "20100315-1234", d.Personnummer)
	assert.Equal(t, "Baddaren 12.55-13.40", d.GroupNameRaw)
	assert.Equal(t, "Deltagare", d.Role)

	require.Len(t, guardians, 1)
	g := guardians[0]
	assert.Equal(t, "20100315-1234", g.MemberPersonnummer)
	assert.Equal(t, "Mamma", g.FirstName)
	assert.Equal(t, "Svensson", g.LastName)
	assert.Equal(t, "070-1111111", g.Phone)
	assert.Equal(t, "mamma@example.com", g.Email)
}

func TestParseWeUnite_SlutDateIgnored(t *testing.T) {
	// Rows with past Slut date must still be collected (spec: Slut date is ignored).
	csv := wuHeader() + "\n" +
		"20100315-1234;Deltagare;Baddaren 08.00-08.55;2024-09-01;2025-06-01;;;;;;;;;;;;;;\n"

	deltagare, _, _, _, err := ParseWeUniteReader(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, deltagare, 1, "Slut date must be ignored — all Deltagare rows included")
}

func TestParseWeUnite_LedareClassified(t *testing.T) {
	csv := wuHeader() + "\n" +
		"19850505-0000;Ledare;Simskola A;2025-09-01;2026-06-01;;;;;;;;;;;;;;\n" +
		"19900101-1111;Huvudledare;A-gruppen;2025-09-01;2026-06-01;;;;;;;;;;;;;;\n"

	deltagare, instructors, _, _, err := ParseWeUniteReader(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Empty(t, deltagare)
	require.Len(t, instructors, 2)
	assert.Equal(t, "Ledare", instructors[0].Role)
	assert.Equal(t, "Huvudledare", instructors[1].Role)
}

func TestParseWeUnite_ThreeGuardianSlots(t *testing.T) {
	csv := wuHeader() + "\n" +
		"20120601-2222;Deltagare;Guldfisken 09.00-09.45;2025-09-01;2026-06-01;" +
		"Pappa;Nilsson;070-2222222;pappa@example.com;" +
		"Mamma;Nilsson;070-3333333;mamma@example.com;" +
		"Farmor;Nilsson;070-4444444;farmor@example.com\n"

	_, _, guardians, _, err := ParseWeUniteReader(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, guardians, 3)
	assert.Equal(t, "Pappa", guardians[0].FirstName)
	assert.Equal(t, "Mamma", guardians[1].FirstName)
	assert.Equal(t, "Farmor", guardians[2].FirstName)
}

func TestParseWeUnite_EmptyGuardianSlotsSkipped(t *testing.T) {
	// Only first guardian slot populated.
	csv := wuHeader() + "\n" +
		"20120601-2222;Deltagare;Guldfisken 09.00-09.45;2025-09-01;2026-06-01;" +
		"Pappa;Nilsson;070-2222222;pappa@example.com;" +
		";;;;;\n"

	_, _, guardians, _, err := ParseWeUniteReader(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, guardians, 1, "empty guardian slots must be skipped")
}

func TestParseWeUnite_SemicolonDelimiter(t *testing.T) {
	// Verify semicolon is used as delimiter (not comma).
	csv := wuHeader() + "\n" +
		"20130707-3333;Deltagare;Masters;2025-09-01;2026-06-01;;;;;;;;;;;;;;\n"

	deltagare, _, _, _, err := ParseWeUniteReader(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, deltagare, 1)
	assert.Equal(t, "Masters", deltagare[0].GroupNameRaw)
}
