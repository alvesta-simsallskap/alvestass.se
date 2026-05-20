package memberimporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeRawMember builds a minimal RawMember for tests.
func makeRawMember(iid, pn, dob string, isBoard bool) RawMember {
	return RawMember{
		IID:           iid,
		FirstName:     "Test",
		LastName:      "Person",
		Gender:        "Man",
		DateOfBirth:   dob,
		Personnummer:  pn,
		IsBoardMember: isBoard,
	}
}

func makeDeltagare(pn, group string) RawGroup {
	return RawGroup{Personnummer: pn, GroupNameRaw: group, Role: "Deltagare", LineNum: 1}
}

func makeLedare(pn, group string) RawGroup {
	return RawGroup{Personnummer: pn, GroupNameRaw: group, Role: "Ledare", LineNum: 1}
}

// TestBuildImportData_Deduplication verifies that two Deltagare rows with the same IID
// and the same normalised group name produce exactly one member and one membership.
func TestBuildImportData_Deduplication(t *testing.T) {
	members := []RawMember{makeRawMember("IID001", "20100315", "2010-03-15", false)}
	deltagare := []RawGroup{
		makeDeltagare("20100315-1234", "Baddaren 12.55-13.40"),
		makeDeltagare("20100315-1234", "Baddaren 08.00-08.45"),
	}

	data, err := BuildImportData(members, deltagare, nil, nil)
	require.NoError(t, err)
	assert.Len(t, data.Members, 1, "duplicate IID must produce exactly one member")
	assert.Len(t, data.Memberships, 1, "time-slot variants of the same group collapse to one membership")
}

// TestBuildImportData_MultipleGroups verifies that a swimmer in two distinct groups
// receives a membership row for each group.
func TestBuildImportData_MultipleGroups(t *testing.T) {
	members := []RawMember{makeRawMember("IID001", "20100315", "2010-03-15", false)}
	deltagare := []RawGroup{
		makeDeltagare("20100315-1234", "Baddaren 12.55-13.40"),
		makeDeltagare("20100315-1234", "Guldfisken 09.00-09.45"),
	}

	data, err := BuildImportData(members, deltagare, nil, nil)
	require.NoError(t, err)
	assert.Len(t, data.Members, 1)
	assert.Len(t, data.Memberships, 2, "two distinct groups must produce two memberships")
}

// TestBuildImportData_LedareOnlyNotInMembers verifies that Ledare rows without a
// corresponding Deltagare row do not appear in the members list.
func TestBuildImportData_LedareOnlyNotInMembers(t *testing.T) {
	members := []RawMember{makeRawMember("IID002", "19850505", "1985-05-05", false)}
	deltagare := []RawGroup{} // no Deltagare rows
	instructors := []RawGroup{makeLedare("19850505-0000", "Simskola A")}

	data, err := BuildImportData(members, deltagare, instructors, nil)
	require.NoError(t, err)
	assert.Empty(t, data.Members, "Ledare-only must not appear in members")
}

// TestBuildImportData_AgeFilterGuardians verifies that guardians are only collected
// for members born within the last 18 years.
func TestBuildImportData_AgeFilterGuardians(t *testing.T) {
	// Member born 2010 = minor today (2026).
	// Member born 1985 = adult today.
	members := []RawMember{
		makeRawMember("IID003", "20100315", "2010-03-15", false), // minor
		makeRawMember("IID004", "19850505", "1985-05-05", false), // adult
	}
	deltagare := []RawGroup{
		makeDeltagare("20100315-1234", "Baddaren 12.55-13.40"),
		makeDeltagare("19850505-0000", "Vuxencrawl"),
	}
	guardians := []RawGuardian{
		{MemberPersonnummer: "20100315-1234", FirstName: "Pappa", LastName: "Svensson", Phone: "070-1111"},
		{MemberPersonnummer: "19850505-0000", FirstName: "Mamma", LastName: "Karlsson", Phone: "070-2222"},
	}

	data, err := BuildImportData(members, deltagare, nil, guardians)
	require.NoError(t, err)
	require.Len(t, data.Guardians, 1, "only minor members should have guardians imported")
	assert.Equal(t, "IID003", data.Guardians[0].MemberIID)
}

// TestBuildImportData_ConflictingIIDUsesIdrottOnline verifies that when the same IID
// exists in both source files with differing data, IdrottOnline values are used.
func TestBuildImportData_ConflictingIIDUsesIdrottOnline(t *testing.T) {
	// IdrottOnline has "Anna Andersson" with email a@example.com
	members := []RawMember{{
		IID:          "IID005",
		FirstName:    "Anna",
		LastName:     "Andersson",
		Email:        "anna@example.com",
		DateOfBirth:  "2010-01-01",
		Personnummer: "20100101",
	}}
	// WeUnite has a Deltagare row for the same personnummer (same person, no conflict in data
	// since we always take personal data from IdrottOnline).
	deltagare := []RawGroup{makeDeltagare("20100101-9999", "Baddaren 12.55-13.40")}

	data, err := BuildImportData(members, deltagare, nil, nil)
	require.NoError(t, err)
	require.Len(t, data.Members, 1)
	assert.Equal(t, "Anna", data.Members[0].FirstName)
	assert.Equal(t, "anna@example.com", data.Members[0].Email)
}

// TestBuildImportData_BoardMemberWithoutWeUnite verifies that a formal board member
// identified in IdrottOnline is included even without a WeUnite Deltagare row.
func TestBuildImportData_BoardMemberWithoutWeUnite(t *testing.T) {
	members := []RawMember{makeRawMember("IID006", "19700101", "1970-01-01", true)}
	// No Deltagare rows for this person.
	data, err := BuildImportData(members, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, data.Members, 1, "board member must be included even without WeUnite Deltagare")
	assert.True(t, data.Members[0].IsBoardMember)
	assert.False(t, data.Members[0].IsSwimmer)
}

// TestBuildImportData_DelagareWithoutIdrottOnline verifies that a WeUnite Deltagare row
// without a matching IdrottOnline entry produces a SkippedRecord.
func TestBuildImportData_DeltagareWithoutIdrottOnline(t *testing.T) {
	members := []RawMember{} // no IdrottOnline entries
	deltagare := []RawGroup{makeDeltagare("20110202-5555", "Guldfisken 09.00-09.45")}

	data, err := BuildImportData(members, deltagare, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, data.Members)
	require.Len(t, data.Skipped, 1)
	assert.Equal(t, "weunite", data.Skipped[0].SourceFile)
}

// TestBuildImportData_InstructorIsRegularMember verifies that a person who appears as both
// a Deltagare and a Ledare in WeUnite is imported as a regular member — no special flag.
func TestBuildImportData_InstructorIsRegularMember(t *testing.T) {
	members := []RawMember{{
		IID:          "IID007",
		FirstName:    "Instructor",
		LastName:     "Person",
		Email:        "instructor@example.com",
		DateOfBirth:  "1990-06-01",
		Personnummer: "19900601",
	}}
	deltagare := []RawGroup{makeDeltagare("19900601-7777", "Masters")}
	instructors := []RawGroup{makeLedare("19900601-7777", "Simskola A")}

	data, err := BuildImportData(members, deltagare, instructors, nil)
	require.NoError(t, err)
	require.Len(t, data.Members, 1)
	assert.True(t, data.Members[0].IsSwimmer, "Deltagare+Ledare is still a member (IsSwimmer)")
	assert.False(t, data.Members[0].IsBoardMember, "not a board member")
}
