# Implementation Plan: CSV Import of Time Report Sessions

**Branch**: `006-import-sessions-csv` | **Date**: 2026-04-26 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/006-import-sessions-csv/spec.md`

## Summary

Add a "Importera tidrapportpass" option to the admin CLI main menu. The administrator selects a local CSV file; the CLI validates all rows, computes an insert/update/skip diff against existing Trailbase records, shows a pre-confirmation summary, and — after explicit confirmation — applies the changes row-by-row via the Trailbase SDK. The composite key `(month_key, training_group, date, title)` determines whether each row is an insert, an update, or a skip.

## Technical Context

**Language/Version**: Go 1.24.4 (see `tools/admin-cli/go.mod`)
**Primary Dependencies**: `github.com/charmbracelet/bubbletea` v1.1.1, `github.com/charmbracelet/bubbles` v0.20.0, `github.com/charmbracelet/lipgloss` v1.0.0, `github.com/trailbaseio/trailbase/client/go/trailbase` (Trailbase SDK), `encoding/csv` (stdlib)
**Storage**: Trailbase SQLite on fly.io (arn) — `time_report_sessions` table (existing, no migration needed)
**Testing**: `go test` + `github.com/stretchr/testify` (already in `go.mod`); unit tests for `internal/importer/`
**Target Platform**: macOS (Apple Silicon + Intel) and Windows 10/11 64-bit
**Project Type**: CLI (admin tool, Go)
**Performance Goals**: Full import of 100 rows in under 30 seconds end-to-end (dominated by network round-trips)
**Constraints**: No schema changes; no new Go module dependencies; all user-visible text in Swedish
**Scale/Scope**: Typical import: 20–100 rows per month; single administrator

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I — Code Quality | ✅ Pass | Typed structs throughout; no `any`; Go vet must pass |
| II — Testing | ✅ Pass | CSV parsing + validation covered by unit tests in `internal/importer/`; no business-logic file without tests |
| III — UX Consistency | ✅ Pass | CLI only; no web UI changes; Swedish text throughout |
| IV — Performance | N/A | CLI tool; Lighthouse not applicable |
| V — Trailbase as sole backend | ✅ Pass | All reads/writes via Trailbase SDK; existing `time_report_sessions` table; no new schema changes |
| VI — Idrottsarenan | N/A | Not involved |
| VII — GDPR | ✅ Pass | `time_report_sessions` contains no personal data (migration comment: "No personal data — training schedule only"); no GDPR gate triggered |

**No gate violations.** Safe to proceed.

## Project Structure

### Documentation (this feature)

```text
specs/006-import-sessions-csv/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── csv-format.md   # CSV file format contract
└── tasks.md             # Phase 2 output (/speckit.tasks — not yet created)
```

### Source Code Changes

```text
tools/admin-cli/
├── cmd/alvestass-admin/
│   └── main.go                          MODIFY — add MenuImport case
├── internal/
│   ├── importer/
│   │   ├── csv.go                       NEW — ParseCSV, validateRow, ComputeDiff
│   │   └── csv_test.go                  NEW — unit tests
│   ├── trailbase/
│   │   ├── client.go                    UNCHANGED
│   │   └── sessions.go                  NEW — TimeReportSession type, ListSessions, CreateSession, UpdateSession
│   └── ui/
│       ├── import.go                    NEW — importModel Bubbletea TUI
│       ├── menu.go                      MODIFY — add MenuImport item + key binding
│       └── help.go                      MODIFY — document new menu option
```

**Structure Decision**: The existing single-module layout under `tools/admin-cli/` is extended in place. New packages (`internal/importer/`, `internal/trailbase/sessions.go`) follow the same conventions as existing code.

## Implementation Notes

### `internal/importer/csv.go`

Package has **zero project-level imports** (stdlib only). All shared types live here to avoid circular imports.

- `ExistingSession` struct — minimal representation of a stored session for diff computation; populated by the caller from `[]TimeReportSession` returned by `ListAllSessions`
- `ImportResult` struct — `{Inserted, Updated, Skipped int}`
- `ParseCSV(path string) ([]SessionRow, []ParseError, error)` — opens file; validates header columns (R-07, returns immediately if any required column missing); parses each row; returns non-nil `error` only for I/O failures
- `ComputeDiff(rows []SessionRow, existing []ExistingSession) ImportDiff` — insert/update/skip classification in memory; no network calls; duplicate composite keys within file: last occurrence wins
- `validateRow(lineNum int, row SessionRow) []ParseError` — per-row rules R-01 through R-06 only (R-07 handled at header stage)
- Validation uses `allowedGroups` set

### `internal/trailbase/sessions.go`

Imports `importer` package (one-way dependency: `trailbase` → `importer`; `importer` never imports `trailbase`).

- `TimeReportSession` struct — internal SDK type for JSON marshalling; not exposed to callers
- `ListAllSessions(monthKeys []string) ([]importer.ExistingSession, error)` — fetches all sessions via `RecordApi[TimeReportSession].List` with cursor-based pagination; maps results to `[]importer.ExistingSession`
- `CreateSession(row importer.SessionRow) error`
- `UpdateSession(id int64, row importer.SessionRow) error`
- `ApplyImport(diff importer.ImportDiff) (importer.ImportResult, error)` — applies inserts first, then updates; returns counts

### `internal/ui/import.go`

Bubbletea model phases (same pattern as `update.go`):
1. `importPhaseInput` — textinput for file path
2. `importPhaseFetching` — spinner while fetching existing sessions
3. `importPhaseSummary` — shows diff counts + confirmation prompt (j/n)
4. `importPhaseApplying` — spinner while applying changes
5. `importPhaseDone` — success message
6. `importPhaseCancelled` / `importPhaseError` — error / cancellation

### `internal/ui/menu.go`

- Add `MenuImport MenuChoice = 4` (shift Quit to 5)
- Add `"Importera tidrapportpass"` to `menuItems` at index 3
- Add key binding `"4"` → `MenuImport`, `"5"` → `MenuQuit`

### `cmd/alvestass-admin/main.go`

- Add `case ui.MenuImport:` dispatching to `ui.RunImport(client)`
