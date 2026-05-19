package memberimporter

// buildFamilyGroups groups imported members by their IdrottOnline FamilyLabel.
// Only groups with 2 or more members are included (a single member cannot form a family).
// Members with empty or unique FamilyLabel values are skipped.
func buildFamilyGroups(rawMembers []RawMember, membersByIID map[string]*ProcessedMember) []FamilyGroup {
	// Build label → []IID map, restricted to members that were actually imported.
	labelToIIDs := make(map[string][]string)
	for i := range rawMembers {
		m := &rawMembers[i]
		if m.FamilyLabel == "" {
			continue
		}
		if _, imported := membersByIID[m.IID]; !imported {
			continue
		}
		labelToIIDs[m.FamilyLabel] = append(labelToIIDs[m.FamilyLabel], m.IID)
	}

	var groups []FamilyGroup
	for label, iids := range labelToIIDs {
		if len(iids) < 2 {
			continue // single-member "family" — skip
		}
		groups = append(groups, FamilyGroup{
			SourceLabel: label,
			MemberIIDs:  iids,
		})
	}
	return groups
}
