package memberimporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

var boardMemberRoles = map[string]bool{
	"Styrelseledamot": true,
	"Ordförande":      true,
	"Vice ordförande": true,
	"Kassör":          true,
	"Sekreterare":     true,
}

// ParseIdrottOnline reads and parses the IdrottOnline export CSV at the given path.
// Returns parsed members and any skipped records.
func ParseIdrottOnline(path string) ([]RawMember, []SkippedRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("kunde inte öppna IdrottOnline-filen: %w", err)
	}
	defer f.Close()
	return ParseIdrottOnlineReader(f)
}

// ParseIdrottOnlineReader is the testable inner parser.
func ParseIdrottOnlineReader(r io.Reader) ([]RawMember, []SkippedRecord, error) {
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("kunde inte läsa header: %w", err)
	}

	idx := buildColIdx(header)

	// Required columns.
	for _, col := range []string{"IdrottsID", "Förnamn", "Efternamn", "Födelsedat./Personnr."} {
		if _, ok := idx[col]; !ok {
			return nil, nil, fmt.Errorf("obligatorisk kolumn saknas i IdrottOnline-fil: %q", col)
		}
	}

	var members []RawMember
	var skipped []SkippedRecord
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("rad %d: läsfel: %w", lineNum+1, err)
		}
		lineNum++

		// Skip "Till målsman för:" guardian-relationship rows.
		if col, ok := idx["Målsman"]; ok && col < len(record) && strings.TrimSpace(record[col]) != "" {
			continue
		}

		iid := strings.TrimSpace(col(record, idx, "IdrottsID"))
		if iid == "" {
			skipped = append(skipped, SkippedRecord{
				SourceFile: "idrottonline",
				Line:       lineNum,
				Reason:     "Saknar IdrottsID",
			})
			continue
		}

		birth := strings.TrimSpace(col(record, idx, "Födelsedat./Personnr."))
		personnummer := ""
		dob := ""
		if len(birth) >= 8 {
			// IdrottOnline exports as "YYYY-MM-DD"; strip dashes to get YYYYMMDD prefix.
			digits := strings.ReplaceAll(birth, "-", "")
			if len(digits) >= 8 {
				personnummer = digits[:8]
				dob = digits[:4] + "-" + digits[4:6] + "-" + digits[6:8]
			}
		}

		roller := strings.TrimSpace(col(record, idx, "Roller"))
		isBoardMember := hasBoardRole(roller)

		members = append(members, RawMember{
			IID:           iid,
			FirstName:     strings.TrimSpace(col(record, idx, "Förnamn")),
			LastName:      strings.TrimSpace(col(record, idx, "Efternamn")),
			Gender:        strings.TrimSpace(col(record, idx, "Kön")),
			DateOfBirth:   dob,
			City:          strings.TrimSpace(col(record, idx, "Kontaktadress - Postort")),
			MemberSince:   strings.TrimSpace(col(record, idx, "Medlem sedan")),
			Email:         strings.TrimSpace(col(record, idx, "E-post kontakt")),
			Phone:         strings.TrimSpace(col(record, idx, "Telefon mobil")),
			Personnummer:  personnummer,
			FamilyLabel:   strings.TrimSpace(col(record, idx, "Familj")),
			IsBoardMember: isBoardMember,
		})
	}

	return members, skipped, nil
}

func hasBoardRole(roller string) bool {
	for role := range boardMemberRoles {
		if strings.Contains(roller, role) {
			return true
		}
	}
	return false
}

// col safely retrieves a value by column name; returns "" if absent.
func col(record []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(record) {
		return ""
	}
	return record[i]
}

// buildColIdx creates a column-name → index map, stripping BOM and whitespace.
func buildColIdx(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, name := range header {
		name = strings.TrimSpace(strings.TrimPrefix(name, "\xef\xbb\xbf"))
		idx[name] = i
	}
	return idx
}
