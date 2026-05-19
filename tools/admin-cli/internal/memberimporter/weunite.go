package memberimporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseWeUnite reads and parses the WeUnite Grupplista CSV at the given path.
// Returns Deltagare rows, Ledare/Huvudledare rows, guardian data, and skipped records.
func ParseWeUnite(path string) (deltagare []RawGroup, instructors []RawGroup, guardians []RawGuardian, skipped []SkippedRecord, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("kunde inte öppna WeUnite-filen: %w", err)
	}
	defer f.Close()
	return ParseWeUniteReader(f)
}

// ParseWeUniteReader is the testable inner parser.
func ParseWeUniteReader(r io.Reader) (deltagare []RawGroup, instructors []RawGroup, guardians []RawGuardian, skipped []SkippedRecord, err error) {
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("kunde inte läsa header: %w", err)
	}

	idx := buildColIdx(header)

	for _, required := range []string{"Personnummer", "Roll", "Grupp"} {
		if _, ok := idx[required]; !ok {
			return nil, nil, nil, nil, fmt.Errorf("obligatorisk kolumn saknas i WeUnite-fil: %q", required)
		}
	}

	lineNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("rad %d: läsfel: %w", lineNum+1, err)
		}
		lineNum++

		pn := strings.TrimSpace(col(record, idx, "Personnummer"))
		roll := strings.TrimSpace(col(record, idx, "Roll"))
		grupp := strings.TrimSpace(col(record, idx, "Grupp"))

		if pn == "" || grupp == "" {
			skipped = append(skipped, SkippedRecord{
				SourceFile: "weunite",
				Line:       lineNum,
				Reason:     "Saknar personnummer eller grupp",
			})
			continue
		}

		row := RawGroup{
			Personnummer: pn,
			GroupNameRaw: grupp,
			Role:         roll,
			LineNum:      lineNum,
		}

		switch roll {
		case "Deltagare":
			deltagare = append(deltagare, row)
			// Extract guardian slots only for Deltagare rows.
			for i := 1; i <= 3; i++ {
				g := extractGuardian(record, idx, pn, i)
				if g != nil {
					guardians = append(guardians, *g)
				}
			}
		case "Ledare", "Huvudledare":
			instructors = append(instructors, row)
		default:
			skipped = append(skipped, SkippedRecord{
				SourceFile: "weunite",
				Line:       lineNum,
				Reason:     fmt.Sprintf("Okänd roll: %q", roll),
			})
		}
	}

	return deltagare, instructors, guardians, skipped, nil
}

// extractGuardian extracts one guardian slot (1-based index n) from a WeUnite row.
// Returns nil if the slot is empty.
func extractGuardian(record []string, idx map[string]int, memberPN string, n int) *RawGuardian {
	prefix := fmt.Sprintf("Målsman %d, ", n)
	firstName := strings.TrimSpace(col(record, idx, prefix+"Förnamn"))
	lastName := strings.TrimSpace(col(record, idx, prefix+"Efternamn"))
	if firstName == "" && lastName == "" {
		return nil
	}
	return &RawGuardian{
		MemberPersonnummer: memberPN,
		FirstName:          firstName,
		LastName:           lastName,
		Phone:              strings.TrimSpace(col(record, idx, prefix+"Telefon")),
		Email:              strings.TrimSpace(col(record, idx, prefix+"E-post")),
	}
}
