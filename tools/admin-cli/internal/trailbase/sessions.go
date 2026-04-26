package trailbase

import (
	"fmt"

	"github.com/alvestass/admin-cli/internal/importer"
	tb "github.com/trailbaseio/trailbase/client/go/trailbase"
)

// TimeReportSession mirrors the time_report_sessions table for use with the Trailbase SDK.
// Internal to this package — callers receive []importer.ExistingSession from ListAllSessions.
type TimeReportSession struct {
	ID            int    `json:"id,omitempty"`
	MonthKey      string `json:"month_key"`
	TrainingGroup string `json:"training_group"`
	Date          string `json:"date"`
	Title         string `json:"title"`
	Hours         int    `json:"hours"`
	Minutes       int    `json:"minutes"`
}

func sessionToExisting(s TimeReportSession) importer.ExistingSession {
	return importer.ExistingSession{
		ID:            int64(s.ID),
		MonthKey:      s.MonthKey,
		TrainingGroup: s.TrainingGroup,
		Date:          s.Date,
		Title:         s.Title,
		Hours:         s.Hours,
		Minutes:       s.Minutes,
	}
}

// CreateSession inserts a new time_report_sessions record.
func (c *Client) CreateSession(row importer.SessionRow) error {
	api := tb.NewRecordApi[TimeReportSession](c.sdk, "time_report_sessions")
	s := TimeReportSession{
		MonthKey:      row.MonthKey,
		TrainingGroup: row.TrainingGroup,
		Date:          row.Date,
		Title:         row.Title,
		Hours:         row.Hours,
		Minutes:       row.Minutes,
	}
	if _, err := api.Create(s); err != nil {
		return fmt.Errorf("kunde inte skapa session %q: %w", row.Title, err)
	}
	return nil
}

// UpdateSession overwrites an existing time_report_sessions record by ID.
func (c *Client) UpdateSession(id int64, row importer.SessionRow) error {
	api := tb.NewRecordApi[TimeReportSession](c.sdk, "time_report_sessions")
	s := TimeReportSession{
		MonthKey:      row.MonthKey,
		TrainingGroup: row.TrainingGroup,
		Date:          row.Date,
		Title:         row.Title,
		Hours:         row.Hours,
		Minutes:       row.Minutes,
	}
	if err := api.Update(tb.IntRecordId(id), s); err != nil {
		return fmt.Errorf("kunde inte uppdatera session %q: %w", row.Title, err)
	}
	return nil
}

// ApplyImport applies inserts and updates from a diff, returning the result counts.
// Inserts are applied first, then updates.
func (c *Client) ApplyImport(diff importer.ImportDiff) (importer.ImportResult, error) {
	var result importer.ImportResult
	for _, row := range diff.Inserts {
		if err := c.CreateSession(row); err != nil {
			return result, err
		}
		result.Inserted++
	}
	for _, u := range diff.Updates {
		if err := c.UpdateSession(u.ID, u.Row); err != nil {
			return result, err
		}
		result.Updated++
	}
	result.Skipped = len(diff.Skipped)
	return result, nil
}

// ListAllSessions fetches all time_report_sessions records for the given month keys.
// One API call per unique month_key; cursor-based pagination with page size 1000.
func (c *Client) ListAllSessions(monthKeys []string) ([]importer.ExistingSession, error) {
	api := tb.NewRecordApi[TimeReportSession](c.sdk, "time_report_sessions")

	seen := make(map[string]bool)
	var all []importer.ExistingSession
	limit := uint64(1000)

	for _, mk := range monthKeys {
		if seen[mk] {
			continue
		}
		seen[mk] = true

		var cursor *string
		for {
			args := &tb.ListArguments{
				Filters: []tb.Filter{
					tb.FilterColumn{Column: "month_key", Op: tb.Equal, Value: mk},
				},
				Pagination: tb.Pagination{
					Limit:  &limit,
					Cursor: cursor,
				},
			}
			resp, err := api.List(args)
			if err != nil {
				return nil, fmt.Errorf("kunde inte hämta sessioner för %s: %w", mk, err)
			}
			for _, s := range resp.Records {
				all = append(all, sessionToExisting(s))
			}
			if resp.Cursor == nil {
				break
			}
			cursor = resp.Cursor
		}
	}
	return all, nil
}
