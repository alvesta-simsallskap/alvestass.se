package importer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alvestass/admin-cli/internal/importer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeCSV writes content to a temp file and returns its path.
func writeCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "test.csv")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
	return p
}

// --- ParseCSV: happy path (T011) ---

func TestParseCSV_AllColumnsInOrder(t *testing.T) {
	path := writeCSV(t, "month_key,training_group,date,title,hours,minutes\n"+
		"2026-04,simskola,2026-04-02,Nybörjare,1,30\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Empty(t, errs)
	require.Len(t, rows, 1)
	assert.Equal(t, importer.SessionRow{
		MonthKey: "2026-04", TrainingGroup: "simskola",
		Date: "2026-04-02", Title: "Nybörjare", Hours: 1, Minutes: 30,
	}, rows[0])
}

func TestParseCSV_NonStandardColumnOrder(t *testing.T) {
	path := writeCSV(t, "title,hours,date,training_group,month_key\n"+
		"Morgonträning,2,2026-04-02,tavlingA,2026-04\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Empty(t, errs)
	require.Len(t, rows, 1)
	assert.Equal(t, "tavlingA", rows[0].TrainingGroup)
	assert.Equal(t, 2, rows[0].Hours)
	assert.Equal(t, 0, rows[0].Minutes) // minutes absent, default 0
}

func TestParseCSV_MinutesAbsentDefaultsToZero(t *testing.T) {
	path := writeCSV(t, "month_key,training_group,date,title,hours\n"+
		"2026-04,teknik,2026-04-03,Teknikpass,1\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Empty(t, errs)
	require.Len(t, rows, 1)
	assert.Equal(t, 0, rows[0].Minutes)
}

func TestParseCSV_HeaderOnlyReturnsEmptySlice(t *testing.T) {
	path := writeCSV(t, "month_key,training_group,date,title,hours\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Empty(t, errs)
	assert.Empty(t, rows)
}

func TestParseCSV_ExtraColumnIgnored(t *testing.T) {
	path := writeCSV(t, "month_key,training_group,date,title,hours,minutes,comment\n"+
		"2026-04,masters,2026-04-04,Masterpass,2,0,ignorerad\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Empty(t, errs)
	require.Len(t, rows, 1)
}

func TestParseCSV_DuplicateKeyLastWins(t *testing.T) {
	path := writeCSV(t, "month_key,training_group,date,title,hours,minutes\n"+
		"2026-04,simskola,2026-04-02,Nybörjare,1,0\n"+
		"2026-04,simskola,2026-04-02,Nybörjare,2,30\n") // same key, updated hours
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Empty(t, errs)
	require.Len(t, rows, 1)
	assert.Equal(t, 2, rows[0].Hours)
	assert.Equal(t, 30, rows[0].Minutes)
}

func TestParseCSV_FileNotFound(t *testing.T) {
	_, _, err := importer.ParseCSV("/no/such/file.csv")
	assert.Error(t, err)
}

// --- ComputeDiff: insert classification (T011) ---

func TestComputeDiff_AllInserts(t *testing.T) {
	rows := []importer.SessionRow{
		{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "A", Hours: 1},
		{MonthKey: "2026-04", TrainingGroup: "teknik", Date: "2026-04-03", Title: "B", Hours: 2},
	}
	diff := importer.ComputeDiff(rows, nil)
	assert.Len(t, diff.Inserts, 2)
	assert.Empty(t, diff.Updates)
	assert.Empty(t, diff.Skipped)
}

func TestComputeDiff_NoExistingAllInsert(t *testing.T) {
	rows := []importer.SessionRow{
		{MonthKey: "2026-04", TrainingGroup: "masters", Date: "2026-04-04", Title: "C", Hours: 1},
	}
	diff := importer.ComputeDiff(rows, []importer.ExistingSession{})
	assert.Len(t, diff.Inserts, 1)
}

// --- ComputeDiff: update and skip classification (T014 / US2) ---

func TestComputeDiff_UpdateWhenHoursDiffer(t *testing.T) {
	existing := []importer.ExistingSession{
		{ID: 42, MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "A", Hours: 1, Minutes: 0},
	}
	rows := []importer.SessionRow{
		{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "A", Hours: 2, Minutes: 0},
	}
	diff := importer.ComputeDiff(rows, existing)
	assert.Empty(t, diff.Inserts)
	require.Len(t, diff.Updates, 1)
	assert.Equal(t, int64(42), diff.Updates[0].ID)
	assert.Equal(t, 2, diff.Updates[0].Row.Hours)
	assert.Empty(t, diff.Skipped)
}

func TestComputeDiff_UpdateWhenMinutesDiffer(t *testing.T) {
	existing := []importer.ExistingSession{
		{ID: 7, MonthKey: "2026-04", TrainingGroup: "teknik", Date: "2026-04-03", Title: "B", Hours: 1, Minutes: 0},
	}
	rows := []importer.SessionRow{
		{MonthKey: "2026-04", TrainingGroup: "teknik", Date: "2026-04-03", Title: "B", Hours: 1, Minutes: 45},
	}
	diff := importer.ComputeDiff(rows, existing)
	require.Len(t, diff.Updates, 1)
	assert.Equal(t, 45, diff.Updates[0].Row.Minutes)
}

func TestComputeDiff_SkipWhenIdentical(t *testing.T) {
	existing := []importer.ExistingSession{
		{ID: 3, MonthKey: "2026-04", TrainingGroup: "masters", Date: "2026-04-04", Title: "C", Hours: 2, Minutes: 0},
	}
	rows := []importer.SessionRow{
		{MonthKey: "2026-04", TrainingGroup: "masters", Date: "2026-04-04", Title: "C", Hours: 2, Minutes: 0},
	}
	diff := importer.ComputeDiff(rows, existing)
	assert.Empty(t, diff.Inserts)
	assert.Empty(t, diff.Updates)
	require.Len(t, diff.Skipped, 1)
}

func TestComputeDiff_MixedInsertUpdateSkip(t *testing.T) {
	existing := []importer.ExistingSession{
		{ID: 1, MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "A", Hours: 1, Minutes: 0},
		{ID: 2, MonthKey: "2026-04", TrainingGroup: "teknik", Date: "2026-04-03", Title: "B", Hours: 1, Minutes: 0},
	}
	rows := []importer.SessionRow{
		// insert (new key)
		{MonthKey: "2026-04", TrainingGroup: "masters", Date: "2026-04-04", Title: "C", Hours: 1},
		// update (hours changed)
		{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "A", Hours: 2},
		// skip (identical)
		{MonthKey: "2026-04", TrainingGroup: "teknik", Date: "2026-04-03", Title: "B", Hours: 1},
	}
	diff := importer.ComputeDiff(rows, existing)
	assert.Len(t, diff.Inserts, 1)
	assert.Len(t, diff.Updates, 1)
	assert.Len(t, diff.Skipped, 1)
}

// --- validateRow: all 7 rules (T018 / US3) ---

func TestValidateRow_ValidRow(t *testing.T) {
	row := importer.SessionRow{
		MonthKey: "2026-04", TrainingGroup: "simskola",
		Date: "2026-04-02", Title: "Test", Hours: 1, Minutes: 30,
	}
	errs := importer.ValidateRow(2, row)
	assert.Empty(t, errs)
}

func TestValidateRow_AllSixAllowedGroups(t *testing.T) {
	groups := []string{"simskola", "tavlingA", "tavlingB", "teknik", "masters", "vuxencrawl"}
	for _, g := range groups {
		row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: g, Date: "2026-04-02", Title: "T", Hours: 0, Minutes: 0}
		errs := importer.ValidateRow(2, row)
		assert.Empty(t, errs, "expected no errors for group %q", g)
	}
}

func TestValidateRow_InvalidTrainingGroup(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "ungdomA", Date: "2026-04-02", Title: "T", Hours: 1}
	errs := importer.ValidateRow(2, row)
	require.Len(t, errs, 1)
	assert.Equal(t, "training_group", errs[0].Column)
}

func TestValidateRow_EmptyMonthKey(t *testing.T) {
	row := importer.SessionRow{TrainingGroup: "simskola", Date: "2026-04-02", Title: "T", Hours: 1}
	errs := importer.ValidateRow(2, row)
	require.NotEmpty(t, errs)
	assert.Equal(t, "month_key", errs[0].Column)
}

func TestValidateRow_EmptyDate(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Title: "T", Hours: 1}
	errs := importer.ValidateRow(2, row)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Column, "date")
}

func TestValidateRow_EmptyTitle(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Hours: 1}
	errs := importer.ValidateRow(2, row)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Column, "title")
}

func TestValidateRow_NegativeHours(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "T", Hours: -1}
	errs := importer.ValidateRow(2, row)
	require.NotEmpty(t, errs)
	assert.Equal(t, "hours", errs[0].Column)
}

func TestValidateRow_HoursZeroIsValid(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "T", Hours: 0, Minutes: 0}
	assert.Empty(t, importer.ValidateRow(2, row))
}

func TestValidateRow_MinutesZeroIsValid(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "T", Hours: 1, Minutes: 0}
	assert.Empty(t, importer.ValidateRow(2, row))
}

func TestValidateRow_Minutes59IsValid(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "T", Hours: 1, Minutes: 59}
	assert.Empty(t, importer.ValidateRow(2, row))
}

func TestValidateRow_Minutes60IsInvalid(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "T", Hours: 1, Minutes: 60}
	errs := importer.ValidateRow(2, row)
	require.NotEmpty(t, errs)
	assert.Equal(t, "minutes", errs[0].Column)
}

func TestValidateRow_NegativeMinutes(t *testing.T) {
	row := importer.SessionRow{MonthKey: "2026-04", TrainingGroup: "simskola", Date: "2026-04-02", Title: "T", Hours: 1, Minutes: -1}
	errs := importer.ValidateRow(2, row)
	require.NotEmpty(t, errs)
	assert.Equal(t, "minutes", errs[0].Column)
}

// --- ParseCSV: validation integration (T018 / US3) ---

func TestParseCSV_MissingRequiredColumn(t *testing.T) {
	// No "hours" column
	path := writeCSV(t, "month_key,training_group,date,title\n"+
		"2026-04,simskola,2026-04-02,Test\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Nil(t, rows)
	require.NotEmpty(t, errs)
	assert.Equal(t, 1, errs[0].Line) // header line
	assert.Equal(t, "hours", errs[0].Column)
}

func TestParseCSV_InvalidTrainingGroupReturnsError(t *testing.T) {
	path := writeCSV(t, "month_key,training_group,date,title,hours\n"+
		"2026-04,ungdomA,2026-04-02,Test,1\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Nil(t, rows)
	require.Len(t, errs, 1)
	assert.Equal(t, 2, errs[0].Line)
}

func TestParseCSV_NegativeHoursReturnsError(t *testing.T) {
	path := writeCSV(t, "month_key,training_group,date,title,hours\n"+
		"2026-04,simskola,2026-04-02,Test,-1\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Nil(t, rows)
	require.NotEmpty(t, errs)
}

func TestParseCSV_MultipleErrorsAllReported(t *testing.T) {
	// Two rows with errors: invalid group + negative hours
	path := writeCSV(t, "month_key,training_group,date,title,hours\n"+
		"2026-04,ungdomA,2026-04-02,Test,1\n"+
		"2026-04,simskola,2026-04-03,Annan,-1\n")
	rows, errs, err := importer.ParseCSV(path)
	require.NoError(t, err)
	assert.Nil(t, rows)
	assert.GreaterOrEqual(t, len(errs), 2, "all errors should be reported")
}
