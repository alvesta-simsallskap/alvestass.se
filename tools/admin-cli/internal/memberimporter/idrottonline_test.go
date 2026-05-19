package memberimporter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalHeader returns a valid IdrottOnline CSV header line.
func ioHeader() string {
	return "IdrottsID;Förnamn;Efternamn;Kön;Födelsedat./Personnr.;Kontaktadress - Postort;Medlem sedan;E-post kontakt;Telefon mobil;Roller;Familj;Målsman"
}

func TestParseIdrottOnline_BasicMember(t *testing.T) {
	csv := ioHeader() + "\n" +
		"IID00000001;Anna;Andersson;Kvinna;20100315-1234;Alvesta;2020-01-01;anna@example.com;070-1234567;;FAM1;\n"

	members, skipped, err := ParseIdrottOnlineReader(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Empty(t, skipped)
	require.Len(t, members, 1)

	m := members[0]
	assert.Equal(t, "IID00000001", m.IID)
	assert.Equal(t, "Anna", m.FirstName)
	assert.Equal(t, "Andersson", m.LastName)
	assert.Equal(t, "Kvinna", m.Gender)
	assert.Equal(t, "2010-03-15", m.DateOfBirth)
	assert.Equal(t, "Alvesta", m.City)
	assert.Equal(t, "2020-01-01", m.MemberSince)
	assert.Equal(t, "anna@example.com", m.Email)
	assert.Equal(t, "070-1234567", m.Phone)
	assert.Equal(t, "FAM1", m.FamilyLabel)
	assert.Equal(t, "20100315", m.Personnummer)
	assert.False(t, m.IsBoardMember)
}

func TestParseIdrottOnline_BoardMemberRoles(t *testing.T) {
	boardRoles := []string{
		"Styrelseledamot",
		"Ordförande",
		"Vice ordförande",
		"Kassör",
		"Sekreterare",
	}
	nonBoardRoles := []string{
		"Klubbadministratör",
		"LOK-stödsansvarig",
		"Kontakt dataskydd",
		"Utbildningsansvarig",
	}

	for _, role := range boardRoles {
		t.Run("board_"+role, func(t *testing.T) {
			csv := ioHeader() + "\n" +
				"IID00000002;Bo;Borg;Man;19800101-9999;Alvesta;2010-01-01;bo@example.com;;" + role + ";;\n"
			members, skipped, err := ParseIdrottOnlineReader(strings.NewReader(csv))
			require.NoError(t, err)
			assert.Empty(t, skipped)
			require.Len(t, members, 1)
			assert.True(t, members[0].IsBoardMember, "expected IsBoardMember for role %q", role)
		})
	}

	for _, role := range nonBoardRoles {
		t.Run("non_board_"+role, func(t *testing.T) {
			csv := ioHeader() + "\n" +
				"IID00000003;Carin;Carlsson;Kvinna;19750601-0000;Alvesta;2015-01-01;carin@example.com;;" + role + ";;\n"
			members, skipped, err := ParseIdrottOnlineReader(strings.NewReader(csv))
			require.NoError(t, err)
			assert.Empty(t, skipped)
			require.Len(t, members, 1)
			assert.False(t, members[0].IsBoardMember, "expected NOT IsBoardMember for role %q", role)
		})
	}
}

func TestParseIdrottOnline_SkipsGuardianRelationshipRows(t *testing.T) {
	// Row with non-empty Målsman column is a "Till målsman för:" row and must be skipped.
	csv := ioHeader() + "\n" +
		"IID00000004;David;Davidsson;Man;20120101-5678;Alvesta;2021-01-01;david@example.com;070-9999999;;;\n" +
		"IID00000005;Eva;Eriksson;Kvinna;19850505-0000;Alvesta;2005-01-01;eva@example.com;;;; Till målsman för: IID00000004\n"

	members, skipped, err := ParseIdrottOnlineReader(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Empty(t, skipped)
	require.Len(t, members, 1, "guardian-relationship rows must be skipped")
	assert.Equal(t, "IID00000004", members[0].IID)
}

func TestParseIdrottOnline_SkipsMissingIID(t *testing.T) {
	csv := ioHeader() + "\n" +
		";Fredrik;Fredriksson;Man;20090909-1111;Alvesta;;;;\n"

	members, skipped, err := ParseIdrottOnlineReader(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Empty(t, members)
	require.Len(t, skipped, 1)
	assert.Equal(t, "idrottonline", skipped[0].SourceFile)
}

func TestParseIdrottOnline_PersonnummerExtraction(t *testing.T) {
	// Legacy format: YYYYMMDD-XXXX
	csv := ioHeader() + "\n" +
		"IID00000006;Greta;Gustafsson;Kvinna;20051220-3333;Alvesta;;;;\n"

	members, _, err := ParseIdrottOnlineReader(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, "20051220", members[0].Personnummer, "personnummer must be YYYYMMDD prefix")
	assert.Equal(t, "2005-12-20", members[0].DateOfBirth)
}

func TestParseIdrottOnline_PersonnummerExtractionISODate(t *testing.T) {
	// Production format: IdrottOnline exports Födelsedat./Personnr. as YYYY-MM-DD
	csv := ioHeader() + "\n" +
		"IID00000007;Hans;Hansson;Man;2011-09-16;Alvesta;;;;\n"

	members, _, err := ParseIdrottOnlineReader(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, "20110916", members[0].Personnummer, "YYYY-MM-DD must be normalised to YYYYMMDD for matching")
	assert.Equal(t, "2011-09-16", members[0].DateOfBirth)
}

func TestParseIdrottOnline_EmptyFile(t *testing.T) {
	members, skipped, err := ParseIdrottOnlineReader(strings.NewReader(ioHeader() + "\n"))
	require.NoError(t, err)
	assert.Empty(t, members)
	assert.Empty(t, skipped)
}
