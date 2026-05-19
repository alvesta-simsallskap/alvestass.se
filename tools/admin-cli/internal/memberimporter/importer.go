package memberimporter

import (
	"fmt"
	"time"
)

// BuildImportData parses and merges the two CSV sources into an ImportData value
// ready for writing to Trailbase.  instructorEmails may be nil (disables the
// IsInstructor flag).  rawGuardians may be nil.
func BuildImportData(
	rawMembers []RawMember,
	deltagare []RawGroup,
	instructorGroups []RawGroup,
	rawGuardians []RawGuardian,
) (ImportData, error) {
	return BuildImportDataWithInstructors(rawMembers, deltagare, instructorGroups, rawGuardians, nil)
}

// BuildImportDataWithInstructors is the full version of BuildImportData that also
// accepts a pre-fetched set of known instructor email addresses.
func BuildImportDataWithInstructors(
	rawMembers []RawMember,
	deltagare []RawGroup,
	instructorGroups []RawGroup,
	rawGuardians []RawGuardian,
	instructorEmails map[string]bool,
) (ImportData, error) {
	// Build personnummer→RawMember and iid→RawMember indexes from IdrottOnline.
	pnIndex := make(map[string]*RawMember, len(rawMembers)) // YYYYMMDD prefix → member
	iidIndex := make(map[string]*RawMember, len(rawMembers))
	for i := range rawMembers {
		m := &rawMembers[i]
		if m.Personnummer != "" {
			pnIndex[m.Personnummer] = m
		}
		iidIndex[m.IID] = m
	}

	var skipped []SkippedRecord
	membersByIID := make(map[string]*ProcessedMember)
	// membershipsByIID[iid][groupName] deduplicates time-slot variants of the same group.
	membershipsByIID := make(map[string]map[string]ProcessedMembership)
	// groups encountered, keyed by normalized name.
	groupsSeen := make(map[string]string) // name → category

	// Step 1: Process Deltagare rows → swimmers.
	for _, dg := range deltagare {
		pn8 := personnummerPrefix(dg.Personnummer)
		raw, ok := pnIndex[pn8]
		if !ok {
			skipped = append(skipped, SkippedRecord{
				SourceFile: "weunite",
				Line:       dg.LineNum,
				Reason:     "Saknar matchande IdrottsID i IdrottOnline",
			})
			continue
		}

		normalName := NormalizeGroupName(dg.GroupNameRaw)
		category := GroupCategory(normalName)
		if category == "" {
			// Unknown group name — use "swim_school" as a fallback and warn.
			category = "swim_school"
		}
		groupsSeen[normalName] = category

		if _, exists := membersByIID[raw.IID]; !exists {
			membersByIID[raw.IID] = processedFromRaw(raw, true, false, instructorEmails)
		}
		if membershipsByIID[raw.IID] == nil {
			membershipsByIID[raw.IID] = make(map[string]ProcessedMembership)
		}
		membershipsByIID[raw.IID][normalName] = ProcessedMembership{
			MemberIID: raw.IID,
			GroupName: normalName,
			Role:      "participant",
		}
	}

	// Step 2: Process board members from IdrottOnline.
	for i := range rawMembers {
		m := &rawMembers[i]
		if !m.IsBoardMember {
			continue
		}
		if existing, ok := membersByIID[m.IID]; ok {
			existing.IsBoardMember = true
		} else {
			membersByIID[m.IID] = processedFromRaw(m, false, true, instructorEmails)
		}
	}

	// Step 3: Cross-reference Ledare/Hoofdledare against instructor emails and warn.
	for _, ins := range instructorGroups {
		pn8 := personnummerPrefix(ins.Personnummer)
		raw, ok := pnIndex[pn8]
		if !ok {
			// Instructor not in IdrottOnline — cannot check email.
			continue
		}
		if instructorEmails != nil && !instructorEmails[raw.Email] {
			skipped = append(skipped, SkippedRecord{
				SourceFile: "weunite",
				Line:       ins.LineNum,
				Reason:     fmt.Sprintf("Ledare %s %s finns inte i instruktörstabellen", raw.FirstName, raw.LastName),
			})
		}
	}

	// Step 4: Collect guardians for minor members only.
	// Deduplicate guardians per swimmer by (firstName+lastName+phone).
	guardiansByIID := make(map[string]map[string]ProcessedGuardian) // iid → dedup key → guardian
	for _, g := range rawGuardians {
		pn8 := personnummerPrefix(g.MemberPersonnummer)
		raw, ok := pnIndex[pn8]
		if !ok {
			continue // member was skipped
		}
		member, ok := membersByIID[raw.IID]
		if !ok {
			continue // member was skipped
		}
		if !isMinor(member.DateOfBirth) {
			continue
		}
		if guardiansByIID[raw.IID] == nil {
			guardiansByIID[raw.IID] = make(map[string]ProcessedGuardian)
		}
		key := g.FirstName + "\x00" + g.LastName + "\x00" + g.Phone
		guardiansByIID[raw.IID][key] = ProcessedGuardian{
			MemberIID: raw.IID,
			FirstName: g.FirstName,
			LastName:  g.LastName,
			Phone:     g.Phone,
			Email:     g.Email,
		}
	}

	// Flatten members, memberships, groups, guardians.
	processedMembers := make([]ProcessedMember, 0, len(membersByIID))
	for _, m := range membersByIID {
		processedMembers = append(processedMembers, *m)
	}

	var processedMemberships []ProcessedMembership
	for iid := range membersByIID {
		for _, ms := range membershipsByIID[iid] {
			processedMemberships = append(processedMemberships, ms)
		}
	}

	processedGroups := make([]ProcessedGroup, 0, len(groupsSeen))
	for name, cat := range groupsSeen {
		processedGroups = append(processedGroups, ProcessedGroup{Name: name, Category: cat})
	}

	var processedGuardians []ProcessedGuardian
	for _, byKey := range guardiansByIID {
		for _, g := range byKey {
			processedGuardians = append(processedGuardians, g)
		}
	}

	// Step 5: Build family groups.
	families := buildFamilyGroups(rawMembers, membersByIID)

	// Build preview.
	swimmerCount, boardCount, instructorCount := 0, 0, 0
	for _, m := range processedMembers {
		if m.IsSwimmer {
			swimmerCount++
		}
		if m.IsBoardMember {
			boardCount++
		}
		if m.IsInstructor {
			instructorCount++
		}
	}

	preview := ImportPreview{
		SwimmerCount:     swimmerCount,
		BoardMemberCount: boardCount,
		InstructorCount:  instructorCount,
		GuardianCount:    len(processedGuardians),
		GroupCount:       len(processedGroups),
		FamilyCount:      len(families),
		SkipCount:        len(skipped),
	}

	return ImportData{
		Members:     processedMembers,
		Groups:      processedGroups,
		Guardians:   processedGuardians,
		Memberships: processedMemberships,
		Families:    families,
		Skipped:     skipped,
		Preview:     preview,
	}, nil
}

// processedFromRaw converts a RawMember to a ProcessedMember.
func processedFromRaw(m *RawMember, isSwimmer, isBoardMember bool, instructorEmails map[string]bool) *ProcessedMember {
	isInstructor := instructorEmails != nil && instructorEmails[m.Email]
	return &ProcessedMember{
		IID:           m.IID,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Gender:        m.Gender,
		DateOfBirth:   m.DateOfBirth,
		City:          m.City,
		MemberSince:   m.MemberSince,
		Email:         m.Email,
		Phone:         m.Phone,
		FamilyLabel:   m.FamilyLabel,
		IsSwimmer:     isSwimmer,
		IsInstructor:  isInstructor,
		IsBoardMember: isBoardMember,
	}
}

// personnummerPrefix returns the first 8 characters (YYYYMMDD) of a personnummer string.
func personnummerPrefix(pn string) string {
	pn = cleanPN(pn)
	if len(pn) >= 8 {
		return pn[:8]
	}
	return pn
}

// cleanPN removes non-digit characters that aren't part of YYYYMMDD.
func cleanPN(pn string) string {
	// WeUnite format: "YYYYMMDD-XXXX"; IdrottOnline: same.
	// We only need the first 8 digits.
	return pn
}

// MemberClient is the interface the importer uses to write to Trailbase.
// Defined here so the UI layer can pass the real *trailbase.Client without creating
// a circular import.
type MemberClient interface {
	ListInstructorEmails() (map[string]bool, error)
	ListTrainingGroups() (map[string]int64, error)
	EnsureTrainingGroup(name, category string, existingByName map[string]int64) (int64, error)
	UpsertMember(m ProcessedMember) error
	ReplaceGuardians(memberIID string, guardians []ProcessedGuardian) error
	ReplaceGroupMemberships(memberIID string, memberships []ProcessedMembership, groupIDByName map[string]int64) error
	ReplaceAllFamilies(groups []FamilyGroup) (int, error)
}

// RunImport parses both CSV files, builds import data, and writes everything to Trailbase.
// It returns a MemberImportResult with counts and skipped records.
func RunImport(client MemberClient, idrottOnlinePath, weUnitePath string) (MemberImportResult, error) {
	// Parse source files.
	rawMembers, ioSkipped, err := ParseIdrottOnline(idrottOnlinePath)
	if err != nil {
		return MemberImportResult{}, err
	}
	deltagare, instructorGroups, rawGuardians, wuSkipped, err := ParseWeUnite(weUnitePath)
	if err != nil {
		return MemberImportResult{}, err
	}

	// Fetch instructor emails for cross-reference.
	instructorEmails, err := client.ListInstructorEmails()
	if err != nil {
		return MemberImportResult{}, fmt.Errorf("kunde inte hämta instruktörer: %w", err)
	}

	// Merge and process.
	allSkipped := append(ioSkipped, wuSkipped...)
	data, err := BuildImportDataWithInstructors(rawMembers, deltagare, instructorGroups, rawGuardians, instructorEmails)
	if err != nil {
		return MemberImportResult{}, err
	}
	data.Skipped = append(allSkipped, data.Skipped...)

	return ApplyImport(client, data)
}

// ApplyImport writes a prepared ImportData to Trailbase and returns the result.
func ApplyImport(client MemberClient, data ImportData) (MemberImportResult, error) {
	// Ensure all training groups exist and build name→ID map.
	groupIDByName, err := client.ListTrainingGroups()
	if err != nil {
		return MemberImportResult{}, fmt.Errorf("kunde inte hämta träningsgrupper: %w", err)
	}
	for _, g := range data.Groups {
		if _, err := client.EnsureTrainingGroup(g.Name, g.Category, groupIDByName); err != nil {
			return MemberImportResult{}, err
		}
	}

	// Upsert members.
	swimmerCount, boardCount := 0, 0
	for _, m := range data.Members {
		if err := client.UpsertMember(m); err != nil {
			return MemberImportResult{}, fmt.Errorf("kunde inte importera medlem %s: %w", m.IID, err)
		}
		if m.IsSwimmer {
			swimmerCount++
		}
		if m.IsBoardMember {
			boardCount++
		}
	}

	// Replace guardians.
	guardianCount := 0
	// Collect guardians by member IID.
	guardiansByMember := make(map[string][]ProcessedGuardian)
	for _, g := range data.Guardians {
		guardiansByMember[g.MemberIID] = append(guardiansByMember[g.MemberIID], g)
	}
	for memberIID, guardians := range guardiansByMember {
		if err := client.ReplaceGuardians(memberIID, guardians); err != nil {
			return MemberImportResult{}, fmt.Errorf("kunde inte importera vårdnadshavare för %s: %w", memberIID, err)
		}
		guardianCount += len(guardians)
	}

	// Replace group memberships.
	membershipsByMember := make(map[string][]ProcessedMembership)
	for _, ms := range data.Memberships {
		membershipsByMember[ms.MemberIID] = append(membershipsByMember[ms.MemberIID], ms)
	}
	for memberIID, memberships := range membershipsByMember {
		if err := client.ReplaceGroupMemberships(memberIID, memberships, groupIDByName); err != nil {
			return MemberImportResult{}, fmt.Errorf("kunde inte importera gruppmedlemskap för %s: %w", memberIID, err)
		}
	}

	// Replace all family links.
	familyLinks, err := client.ReplaceAllFamilies(data.Families)
	if err != nil {
		return MemberImportResult{}, fmt.Errorf("kunde inte importera familjekopplingar: %w", err)
	}

	// Build per-category group counts.
	groupsByCategory := make(map[string]int)
	for _, g := range data.Groups {
		groupsByCategory[g.Category]++
	}

	return MemberImportResult{
		MembersImported:      len(data.Members),
		SwimmersImported:     swimmerCount,
		BoardMembersImported: boardCount,
		GuardiansImported:    guardianCount,
		GroupsImported:       len(data.Groups),
		FamilyLinksImported:  familyLinks,
		GroupsByCategory:     groupsByCategory,
		Skipped:              data.Skipped,
	}, nil
}

// isMinor returns true if the member's date of birth indicates they are under 18 today.
func isMinor(dob string) bool {
	if len(dob) < 10 {
		return false
	}
	t, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return false
	}
	return time.Now().AddDate(-18, 0, 0).Before(t)
}
