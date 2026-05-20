package memberimporter

// RawMember is a person parsed from the IdrottOnline CSV export.
// Personnummer is a transient join key — never written to Trailbase.
type RawMember struct {
	IID           string
	FirstName     string
	LastName      string
	Gender        string
	DateOfBirth   string // YYYY-MM-DD
	City          string
	MemberSince   string // YYYY-MM-DD; blank if absent
	Email         string
	Phone         string
	Personnummer  string // YYYYMMDD prefix; transient join key with WeUnite
	FamilyLabel   string // Familj field from IdrottOnline
	IsBoardMember bool
}

// RawGroup is one row from the WeUnite Grupplista (one person × one group booking).
type RawGroup struct {
	Personnummer string // swimmer's personnummer (YYYYMMDD-XXXX)
	GroupNameRaw string // group name before time-slot stripping
	Role         string // "Deltagare", "Ledare", or "Huvudledare"
	LineNum      int    // 1-based source line number for error reporting
}

// RawGuardian holds one guardian slot extracted from a WeUnite row.
type RawGuardian struct {
	MemberPersonnummer string // swimmer's personnummer (for linking to member IID)
	FirstName          string
	LastName           string
	Phone              string
	Email              string
}

// ProcessedMember is a member ready for Trailbase upsert.
// IsSwimmer is a transient flag used for import counting only — not written to the DB.
type ProcessedMember struct {
	IID           string
	FirstName     string
	LastName      string
	Gender        string
	DateOfBirth   string
	City          string
	MemberSince   string
	Email         string
	Phone         string
	FamilyLabel   string
	IsSwimmer     bool // transient: true if sourced from a WeUnite Deltagare row
	IsBoardMember bool
}

// ProcessedGroup is a training group ready for Trailbase upsert (keyed on name).
type ProcessedGroup struct {
	Name     string
	Category string
}

// ProcessedGuardian is a guardian row ready for Trailbase insert.
type ProcessedGuardian struct {
	MemberIID string
	FirstName string
	LastName  string
	Phone     string
	Email     string
}

// ProcessedMembership is a member-training-group join row ready for insert.
type ProcessedMembership struct {
	MemberIID string
	GroupName string
	Role      string // always "participant" for imported Deltagare rows
}

// FamilyGroup represents a family unit: a shared label and the IIDs of its members.
type FamilyGroup struct {
	SourceLabel string
	MemberIIDs  []string
}

// ImportData is the fully processed, Trailbase-ready result of parsing and merging both CSV files.
type ImportData struct {
	Members     []ProcessedMember
	Groups      []ProcessedGroup    // deduplicated by name
	Guardians   []ProcessedGuardian // all guardians (for minor members only)
	Memberships []ProcessedMembership
	Families    []FamilyGroup
	Skipped     []SkippedRecord
	Preview     ImportPreview
}

// ImportPreview is shown before the user confirms the import.
type ImportPreview struct {
	SwimmerCount     int
	BoardMemberCount int
	GuardianCount    int
	GroupCount       int
	FamilyCount      int
	SkipCount        int
}

// MemberImportResult is the final outcome of a completed import run.
type MemberImportResult struct {
	MembersImported      int
	BoardMembersImported int
	GuardiansImported    int
	GroupsImported       int
	FamilyLinksImported  int
	GroupsByCategory     map[string]int
	Skipped              []SkippedRecord
}

// SkippedRecord documents a single record that could not be imported.
type SkippedRecord struct {
	SourceFile string // "idrottonline" or "weunite"
	Line       int
	Reason     string // Swedish human-readable description
}
