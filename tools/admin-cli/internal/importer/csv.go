package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// SessionRow holds one parsed row from the CSV import file.
type SessionRow struct {
	MonthKey      string
	TrainingGroup string
	Date          string
	Title         string
	Hours         int
	Minutes       int
}

// ExistingSession is a session already stored in Trailbase, used by ComputeDiff.
// Defined here so the importer package has no dependency on the trailbase package.
type ExistingSession struct {
	ID            int64
	MonthKey      string
	TrainingGroup string
	Date          string
	Title         string
	Hours         int
	Minutes       int
}

// ParseError describes a single validation problem found while parsing the CSV file.
type ParseError struct {
	Line    int    // 1-based line number (header = line 1)
	Column  string // column name, empty for row-level errors
	Value   string // the offending value
	Message string // Swedish human-readable description
}

// SessionUpdate pairs a CSV row (new values) with the Trailbase record ID to overwrite.
type SessionUpdate struct {
	ID  int64
	Row SessionRow
}

// ImportDiff is the result of comparing CSV rows against existing Trailbase records.
type ImportDiff struct {
	Inserts []SessionRow
	Updates []SessionUpdate
	Skipped []SessionRow
}

// ImportResult is the outcome after applying a confirmed import.
type ImportResult struct {
	Inserted int
	Updated  int
	Skipped  int
}

var allowedGroups = map[string]bool{
	"simskola":   true,
	"tavlingA":   true,
	"tavlingB":   true,
	"teknik":     true,
	"masters":    true,
	"vuxencrawl": true,
}

var requiredColumns = []string{"month_key", "training_group", "date", "title", "hours"}

// ParseCSV reads and parses a CSV file at the given path.
// It returns the parsed rows and any validation errors found.
// A non-nil error indicates an I/O failure (file unreadable); validation
// problems are returned as []ParseError with a nil error.
// If the file contains only a header row (no data rows), an empty slice is returned.
func ParseCSV(path string) ([]SessionRow, []ParseError, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("kunde inte öppna filen: %w", err)
	}
	defer f.Close()

	return parseCSVReader(f)
}

func parseCSVReader(r io.Reader) ([]SessionRow, []ParseError, error) {
	reader := csv.NewReader(r)

	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil, nil // completely empty file
	}
	if err != nil {
		return nil, nil, fmt.Errorf("kunde inte läsa CSV-filen: %w", err)
	}

	// Build column index map (trimming BOM / whitespace from header names).
	colIdx := make(map[string]int, len(header))
	for i, name := range header {
		colIdx[strings.TrimSpace(strings.TrimPrefix(name, "\xef\xbb\xbf"))] = i
	}

	// R-07: check required columns (header-level, line 1).
	var parseErrs []ParseError
	for _, col := range requiredColumns {
		if _, ok := colIdx[col]; !ok {
			parseErrs = append(parseErrs, ParseError{
				Line:    1,
				Column:  col,
				Message: fmt.Sprintf("obligatorisk kolumn saknas: '%s'", col),
			})
		}
	}
	if len(parseErrs) > 0 {
		return nil, parseErrs, nil
	}

	_, hasMinutes := colIdx["minutes"]

	// Parse rows — later rows overwrite earlier ones for the same composite key.
	// Collect all validation errors across the entire file before returning.
	rowMap := make(map[string]SessionRow)
	var keyOrder []string
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("kunde inte läsa CSV-filen: %w", err)
		}
		lineNum++

		row := SessionRow{
			MonthKey:      strings.TrimSpace(record[colIdx["month_key"]]),
			TrainingGroup: strings.TrimSpace(record[colIdx["training_group"]]),
			Date:          strings.TrimSpace(record[colIdx["date"]]),
			Title:         strings.TrimSpace(record[colIdx["title"]]),
		}

		hoursStr := strings.TrimSpace(record[colIdx["hours"]])
		if h, err := strconv.Atoi(hoursStr); err == nil {
			row.Hours = h
		} else {
			row.Hours = -1 // triggers R-05 in ValidateRow
		}

		if hasMinutes {
			minStr := strings.TrimSpace(record[colIdx["minutes"]])
			if mn, err := strconv.Atoi(minStr); err == nil {
				row.Minutes = mn
			} else {
				row.Minutes = -1 // triggers R-06 in ValidateRow
			}
		}
		// minutes defaults to 0 when column is absent (zero value)

		// R-01 through R-06: per-row validation.
		if errs := ValidateRow(lineNum, row); len(errs) > 0 {
			parseErrs = append(parseErrs, errs...)
			continue // don't add invalid row to the map
		}

		key := compositeKey(row.MonthKey, row.TrainingGroup, row.Date, row.Title)
		if _, exists := rowMap[key]; !exists {
			keyOrder = append(keyOrder, key)
		}
		rowMap[key] = row
	}

	// Any validation error means we return no rows — caller must not import.
	if len(parseErrs) > 0 {
		return nil, parseErrs, nil
	}

	rows := make([]SessionRow, 0, len(rowMap))
	for _, k := range keyOrder {
		rows = append(rows, rowMap[k])
	}
	return rows, nil, nil
}

// ComputeDiff classifies each CSV row as an insert, update, or skip
// by comparing against existing sessions fetched from Trailbase.
func ComputeDiff(rows []SessionRow, existing []ExistingSession) ImportDiff {
	existingMap := make(map[string]ExistingSession, len(existing))
	for _, e := range existing {
		existingMap[compositeKey(e.MonthKey, e.TrainingGroup, e.Date, e.Title)] = e
	}

	var diff ImportDiff
	for _, row := range rows {
		key := compositeKey(row.MonthKey, row.TrainingGroup, row.Date, row.Title)
		e, exists := existingMap[key]
		if !exists {
			diff.Inserts = append(diff.Inserts, row)
		} else if e.Hours != row.Hours || e.Minutes != row.Minutes {
			diff.Updates = append(diff.Updates, SessionUpdate{ID: e.ID, Row: row})
		} else {
			diff.Skipped = append(diff.Skipped, row)
		}
	}
	return diff
}

// ValidateRow checks per-row field rules R-01 through R-06.
// R-07 (required header columns) is checked once at header-parse time in ParseCSV.
func ValidateRow(lineNum int, row SessionRow) []ParseError {
	var errs []ParseError

	if row.MonthKey == "" {
		errs = append(errs, ParseError{Line: lineNum, Column: "month_key",
			Message: "month_key får inte vara tomt"})
	}
	if !allowedGroups[row.TrainingGroup] {
		errs = append(errs, ParseError{Line: lineNum, Column: "training_group",
			Value:   row.TrainingGroup,
			Message: fmt.Sprintf("träningsgrupp '%s' är inte giltig", row.TrainingGroup)})
	}
	if row.Date == "" {
		errs = append(errs, ParseError{Line: lineNum, Column: "date",
			Message: "datum får inte vara tomt"})
	}
	if row.Title == "" {
		errs = append(errs, ParseError{Line: lineNum, Column: "title",
			Message: "titel får inte vara tomt"})
	}
	if row.Hours < 0 {
		errs = append(errs, ParseError{Line: lineNum, Column: "hours",
			Value:   strconv.Itoa(row.Hours),
			Message: "timmar måste vara ett heltal ≥ 0"})
	}
	if row.Minutes < 0 || row.Minutes >= 60 {
		errs = append(errs, ParseError{Line: lineNum, Column: "minutes",
			Value:   strconv.Itoa(row.Minutes),
			Message: "minuter måste vara ett heltal mellan 0 och 59"})
	}
	return errs
}

func compositeKey(monthKey, trainingGroup, date, title string) string {
	return monthKey + "\x00" + trainingGroup + "\x00" + date + "\x00" + title
}
