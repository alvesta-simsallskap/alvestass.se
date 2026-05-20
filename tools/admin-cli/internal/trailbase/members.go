package trailbase

import (
	"fmt"
	"strconv"

	"github.com/alvestass/admin-cli/internal/memberimporter"
	tb "github.com/trailbaseio/trailbase/client/go/trailbase"
)

// ---- Table row types --------------------------------------------------------

// MemberRow mirrors the members table.
type MemberRow struct {
	ID            int64  `json:"id,omitempty"`
	IID           string `json:"iid"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Gender        string `json:"gender,omitempty"`
	DateOfBirth   string `json:"date_of_birth,omitempty"`
	City          string `json:"city,omitempty"`
	MemberSince   string `json:"member_since,omitempty"`
	Email         string `json:"email,omitempty"`
	Phone         string `json:"phone,omitempty"`
	IsBoardMember int    `json:"is_board_member"`
}

// GuardianRow mirrors the guardians table.
type GuardianRow struct {
	ID        int64  `json:"id,omitempty"`
	MemberIID string `json:"member_iid"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
}

// TrainingGroupRow mirrors the training_groups table.
type TrainingGroupRow struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// MemberTrainingGroupRow mirrors the member_training_groups table.
type MemberTrainingGroupRow struct {
	ID        int64  `json:"id,omitempty"`
	MemberIID string `json:"member_iid"`
	GroupID   int64  `json:"group_id"`
	Role      string `json:"role"`
}

// InstructorRow mirrors the instructors table (read-only in this context).
type InstructorRow struct {
	ID             int64  `json:"id,omitempty"`
	Email          string `json:"email"`
	SwimSchoolRate int    `json:"swim_school_rate"`
	CoachRate      *int   `json:"coach_rate,omitempty"`
}

// FamilyRow mirrors the families table.
type FamilyRow struct {
	ID          int64  `json:"id,omitempty"`
	SourceLabel string `json:"source_label,omitempty"`
}

// FamilyMemberRow mirrors the family_members table.
type FamilyMemberRow struct {
	ID        int64  `json:"id,omitempty"`
	FamilyID  int64  `json:"family_id"`
	MemberIID string `json:"member_iid"`
}

// ---- members ----------------------------------------------------------------

// UpsertMember creates a new member record or updates it if the IID already exists.
func (c *Client) UpsertMember(m memberimporter.ProcessedMember) error {
	api := tb.NewRecordApi[MemberRow](c.sdk, "members")
	row := MemberRow{
		IID:           m.IID,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Gender:        m.Gender,
		DateOfBirth:   m.DateOfBirth,
		City:          m.City,
		MemberSince:   m.MemberSince,
		Email:         m.Email,
		Phone:         m.Phone,
		IsBoardMember: boolToInt(m.IsBoardMember),
	}
	_, err := api.Create(row)
	if err == nil {
		return nil
	}
	// Conflict: look up the integer PK by iid, then update.
	limit := uint64(1)
	resp, listErr := api.List(&tb.ListArguments{
		Filters:    []tb.Filter{tb.FilterColumn{Column: "iid", Op: tb.Equal, Value: m.IID}},
		Pagination: tb.Pagination{Limit: &limit},
	})
	if listErr != nil || len(resp.Records) == 0 {
		return fmt.Errorf("kunde inte hitta befintlig post för IID %s: %w", m.IID, err)
	}
	return api.Update(tb.IntRecordId(resp.Records[0].ID), row)
}

// ---- training_groups --------------------------------------------------------

// ListTrainingGroups fetches all training_groups and returns a name→ID map.
func (c *Client) ListTrainingGroups() (map[string]int64, error) {
	api := tb.NewRecordApi[TrainingGroupRow](c.sdk, "training_groups")
	limit := uint64(1000)
	result := make(map[string]int64)
	var cursor *string
	for {
		resp, err := api.List(&tb.ListArguments{
			Pagination: tb.Pagination{Limit: &limit, Cursor: cursor},
		})
		if err != nil {
			return nil, fmt.Errorf("kunde inte hämta träningsgrupper: %w", err)
		}
		for _, g := range resp.Records {
			result[g.Name] = g.ID
		}
		if resp.Cursor == nil {
			break
		}
		cursor = resp.Cursor
	}
	return result, nil
}

// EnsureTrainingGroup returns the group ID for the given name, creating the record
// if it does not exist. existingByName is updated in place on creation.
func (c *Client) EnsureTrainingGroup(name, category string, existingByName map[string]int64) (int64, error) {
	if id, ok := existingByName[name]; ok {
		return id, nil
	}
	api := tb.NewRecordApi[TrainingGroupRow](c.sdk, "training_groups")
	rid, err := api.Create(TrainingGroupRow{Name: name, Category: category})
	if err != nil {
		return 0, fmt.Errorf("kunde inte skapa träningsgrupp %q: %w", name, err)
	}
	id, err := strconv.ParseInt(rid.ToString(), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("oväntat ID från Trailbase för grupp %q: %w", name, err)
	}
	existingByName[name] = id
	return id, nil
}

// ---- guardians --------------------------------------------------------------

// ReplaceGuardians deletes all existing guardian rows for memberIID and inserts the new ones.
func (c *Client) ReplaceGuardians(memberIID string, guardians []memberimporter.ProcessedGuardian) error {
	api := tb.NewRecordApi[GuardianRow](c.sdk, "guardians")
	limit := uint64(1000)

	// Delete all existing guardians for this member.
	resp, err := api.List(&tb.ListArguments{
		Filters:    []tb.Filter{tb.FilterColumn{Column: "member_iid", Op: tb.Equal, Value: memberIID}},
		Pagination: tb.Pagination{Limit: &limit},
	})
	if err != nil {
		return fmt.Errorf("kunde inte hämta befintliga vårdnadshavare för %s: %w", memberIID, err)
	}
	for _, g := range resp.Records {
		if err := api.Delete(tb.IntRecordId(g.ID)); err != nil {
			return fmt.Errorf("kunde inte radera vårdnadshavare %d: %w", g.ID, err)
		}
	}

	// Insert new guardians.
	for _, g := range guardians {
		row := GuardianRow{
			MemberIID: memberIID,
			FirstName: g.FirstName,
			LastName:  g.LastName,
			Phone:     g.Phone,
			Email:     g.Email,
		}
		if _, err := api.Create(row); err != nil {
			return fmt.Errorf("kunde inte skapa vårdnadshavare för %s: %w", memberIID, err)
		}
	}
	return nil
}

// ---- member_training_groups -------------------------------------------------

// ReplaceGroupMemberships deletes all existing group memberships for memberIID and inserts new ones.
// groupIDByName must map normalised group names to Trailbase group IDs.
func (c *Client) ReplaceGroupMemberships(memberIID string, memberships []memberimporter.ProcessedMembership, groupIDByName map[string]int64) error {
	api := tb.NewRecordApi[MemberTrainingGroupRow](c.sdk, "member_training_groups")
	limit := uint64(1000)

	// Delete all existing memberships for this member.
	resp, err := api.List(&tb.ListArguments{
		Filters:    []tb.Filter{tb.FilterColumn{Column: "member_iid", Op: tb.Equal, Value: memberIID}},
		Pagination: tb.Pagination{Limit: &limit},
	})
	if err != nil {
		return fmt.Errorf("kunde inte hämta befintliga gruppmedlemskap för %s: %w", memberIID, err)
	}
	for _, m := range resp.Records {
		if err := api.Delete(tb.IntRecordId(m.ID)); err != nil {
			return fmt.Errorf("kunde inte radera gruppmedlemskap %d: %w", m.ID, err)
		}
	}

	// Insert new memberships.
	for _, ms := range memberships {
		gid, ok := groupIDByName[ms.GroupName]
		if !ok {
			return fmt.Errorf("träningsgrupp %q hittades inte i Trailbase (intern fel)", ms.GroupName)
		}
		row := MemberTrainingGroupRow{
			MemberIID: memberIID,
			GroupID:   gid,
			Role:      ms.Role,
		}
		if _, err := api.Create(row); err != nil {
			return fmt.Errorf("kunde inte skapa gruppmedlemskap för %s i grupp %q: %w", memberIID, ms.GroupName, err)
		}
	}
	return nil
}

// ---- families ---------------------------------------------------------------

// ReplaceAllFamilies deletes all existing family data and inserts fresh family groups.
// Returns the total number of family_members rows inserted.
func (c *Client) ReplaceAllFamilies(groups []memberimporter.FamilyGroup) (int, error) {
	fmAPI := tb.NewRecordApi[FamilyMemberRow](c.sdk, "family_members")
	fAPI := tb.NewRecordApi[FamilyRow](c.sdk, "families")
	limit := uint64(1000)

	// Delete all family_members first (avoids FK constraint issues).
	for {
		resp, err := fmAPI.List(&tb.ListArguments{Pagination: tb.Pagination{Limit: &limit}})
		if err != nil {
			return 0, fmt.Errorf("kunde inte hämta familjemedlemmar: %w", err)
		}
		if len(resp.Records) == 0 {
			break
		}
		for _, fm := range resp.Records {
			if err := fmAPI.Delete(tb.IntRecordId(fm.ID)); err != nil {
				return 0, fmt.Errorf("kunde inte radera familjemedlem %d: %w", fm.ID, err)
			}
		}
	}

	// Delete all families.
	for {
		resp, err := fAPI.List(&tb.ListArguments{Pagination: tb.Pagination{Limit: &limit}})
		if err != nil {
			return 0, fmt.Errorf("kunde inte hämta familjer: %w", err)
		}
		if len(resp.Records) == 0 {
			break
		}
		for _, f := range resp.Records {
			if err := fAPI.Delete(tb.IntRecordId(f.ID)); err != nil {
				return 0, fmt.Errorf("kunde inte radera familj %d: %w", f.ID, err)
			}
		}
	}

	// Insert new families and their member links.
	links := 0
	for _, group := range groups {
		rid, err := fAPI.Create(FamilyRow{SourceLabel: group.SourceLabel})
		if err != nil {
			return links, fmt.Errorf("kunde inte skapa familj %q: %w", group.SourceLabel, err)
		}
		fid, err := strconv.ParseInt(rid.ToString(), 10, 64)
		if err != nil {
			return links, fmt.Errorf("oväntat ID från Trailbase för familj %q: %w", group.SourceLabel, err)
		}
		for _, iid := range group.MemberIIDs {
			if _, err := fmAPI.Create(FamilyMemberRow{FamilyID: fid, MemberIID: iid}); err != nil {
				return links, fmt.Errorf("kunde inte skapa familjemedlemslänk %s→%d: %w", iid, fid, err)
			}
			links++
		}
	}
	return links, nil
}

// ---- instructors (read-only) ------------------------------------------------

// ListInstructorEmails returns the set of all email addresses in the instructors table.
func (c *Client) ListInstructorEmails() (map[string]bool, error) {
	api := tb.NewRecordApi[InstructorRow](c.sdk, "instructors")
	limit := uint64(1000)
	result := make(map[string]bool)
	var cursor *string
	for {
		resp, err := api.List(&tb.ListArguments{
			Pagination: tb.Pagination{Limit: &limit, Cursor: cursor},
		})
		if err != nil {
			return nil, fmt.Errorf("kunde inte hämta instruktörer: %w", err)
		}
		for _, ins := range resp.Records {
			result[ins.Email] = true
		}
		if resp.Cursor == nil {
			break
		}
		cursor = resp.Cursor
	}
	return result, nil
}

// ---- helpers ----------------------------------------------------------------

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
