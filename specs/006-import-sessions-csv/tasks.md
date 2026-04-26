# Tasks: CSV Import of Time Report Sessions

**Input**: Design documents from `specs/006-import-sessions-csv/`
**Branch**: `006-import-sessions-csv`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: Required — `internal/importer/` is a business-logic module; Principle II of the project constitution mandates unit tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths included in every description

---

## Phase 1: Setup (Scaffold New Files)

**Purpose**: Create the new files so all downstream tasks have clear targets. The existing project builds clean; this phase only adds empty skeletons.

- [ ] T001 [P] Scaffold `tools/admin-cli/internal/importer/csv.go` with package declaration and empty type stubs: `SessionRow`, `ExistingSession`, `ParseError`, `ImportDiff`, `SessionUpdate`, `ImportResult` (all fields present per data-model.md; no method bodies yet; zero project-level imports)
- [ ] T002 [P] Scaffold `tools/admin-cli/internal/trailbase/sessions.go` with package declaration, `import "github.com/alvestass/admin-cli/internal/importer"`, and empty type stub: `TimeReportSession` only (all fields present per data-model.md; `ImportResult` lives in `importer`, not here)
- [ ] T003 [P] Scaffold `tools/admin-cli/internal/importer/csv_test.go` with package declaration and a single placeholder `TestPlaceholder` that passes, confirming the test binary compiles

**Checkpoint**: `go build ./...` and `go test ./...` MUST pass before Phase 2

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before any user story implementation begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T004 Implement `Client.ListAllSessions(monthKeys []string) ([]importer.ExistingSession, error)` in `tools/admin-cli/internal/trailbase/sessions.go` — fetches all `time_report_sessions` records for the given month keys using `RecordApi[TimeReportSession].List` with a `FilterColumn` on `month_key` and cursor-based pagination (page size 1000, follow cursor until nil); maps each `TimeReportSession` to `importer.ExistingSession` before returning
- [ ] T005 [P] Add `MenuImport MenuChoice = 4` constant, shift `MenuQuit` to `5`, add `"Importera tidrapportpass"` as the fourth item in `menuItems`, and add key bindings `"4"` → `MenuImport` / `"5"` → `MenuQuit` in `tools/admin-cli/internal/ui/menu.go`
- [ ] T006 Add `case ui.MenuImport:` dispatch calling `ui.RunImport(client)` (with error printing identical to the existing `MenuUpdate` case) in `tools/admin-cli/cmd/alvestass-admin/main.go`; add a stub `RunImport(client *trailbase.Client) error` returning `nil` in `tools/admin-cli/internal/ui/import.go` so the project compiles

**Checkpoint**: `go build ./...` MUST pass; running the CLI MUST show five menu items including "Importera tidrapportpass"; selecting it MUST return immediately without error

---

## Phase 3: User Story 1 — Import a New Session Schedule (Priority: P1) 🎯 MVP

**Goal**: An administrator can select a valid CSV file of new sessions, see a summary showing only inserts, confirm, and have all rows inserted into Trailbase.

**Independent Test**: Create a CSV with 3 rows (different training groups, all new). Launch CLI → select import → enter file path → confirm insert summary → verify `go test` passes and (manually) verify the 3 rows appear in Trailbase.

### Implementation for User Story 1

- [ ] T007 [US1] Implement `ParseCSV(path string) ([]SessionRow, []ParseError, error)` happy path in `tools/admin-cli/internal/importer/csv.go`: open file, read header row to build column-index map, parse each data row into `SessionRow`, default `Minutes` to 0 if column absent, return non-nil `error` only for I/O failure (no validation yet — validation is US3)
- [ ] T008 [US1] Implement `ComputeDiff(rows []SessionRow, existing []ExistingSession) ImportDiff` in `tools/admin-cli/internal/importer/csv.go` for **insert classification only**: build a composite-key map from `existing`; for each `SessionRow` where the key `(MonthKey+TrainingGroup+Date+Title)` is absent → append to `Inserts`; last occurrence in `rows` wins for duplicate keys within the file; `importer` package has no imports from `trailbase`
- [ ] T009 [US1] Implement `Client.CreateSession(row importer.SessionRow) error` and `Client.ApplyImport(diff importer.ImportDiff) (importer.ImportResult, error)` (inserts only, updates handled in US2) in `tools/admin-cli/internal/trailbase/sessions.go`
- [ ] T010 [US1] Replace the `RunImport` stub with the full `importModel` Bubbletea model in `tools/admin-cli/internal/ui/import.go` covering phases: `importPhaseInput` (textinput for file path, same pattern as `ui/update.go`), `importPhaseFetching` (spinner + `ListAllSessions` call), `importPhaseSummary` (diff counts + "j/n" prompt), `importPhaseApplying` (spinner + `ApplyImport` call), `importPhaseDone` (success message with counts), `importPhaseCancelled`; **additionally**: after `ComputeDiff`, if all three counts (Inserts, Updates, Skipped) are zero, skip the confirmation prompt and transition directly to `importPhaseDone` displaying "0 rader hittades. Inget att importera."
- [ ] T011 [P] [US1] Unit tests for `ParseCSV` valid input in `tools/admin-cli/internal/importer/csv_test.go`: (a) all 6 columns present in header order, (b) columns in non-standard order, (c) `minutes` column absent → defaults to 0, (d) empty file (header only) → returns empty slice with no errors, (e) extra unknown column silently ignored

**Checkpoint**: US1 fully functional — administrator can import a fresh CSV of 20 sessions in under 30 seconds; `go test ./internal/importer/...` passes green

---

## Phase 4: User Story 2 — Update an Existing Session (Priority: P2)

**Goal**: Re-importing a CSV where some rows match existing sessions classifies those rows as updates (changed values) or skips (identical values), and applies updates to Trailbase.

**Independent Test**: Import a CSV once (US1 checkpoint). Change `hours` for one row. Re-import — CLI shows "0 inserts, 1 update, N skips". Confirm. Verify the record in Trailbase reflects the new hours. Re-import unchanged CSV — CLI shows "0 inserts, 0 updates, N skips".

### Implementation for User Story 2

- [ ] T012 [US2] Extend `ComputeDiff` in `tools/admin-cli/internal/importer/csv.go` to classify rows against existing data: key found AND `Hours`/`Minutes` differ → append to `Updates` (as `SessionUpdate{ID: existing.ID, Row: row}`); key found AND values identical → append to `Skipped`
- [ ] T013 [US2] Implement `Client.UpdateSession(id int64, row importer.SessionRow) error` and extend `Client.ApplyImport(diff importer.ImportDiff) (importer.ImportResult, error)` to apply updates (via `UpdateSession`) after inserts in `tools/admin-cli/internal/trailbase/sessions.go`
- [ ] T014 [P] [US2] Unit tests for `ComputeDiff` update and skip classification in `tools/admin-cli/internal/importer/csv_test.go`: (a) key exists, hours differ → update; (b) key exists, minutes differ → update; (c) key exists, identical values → skip; (d) mix of insert + update + skip in one call

**Checkpoint**: US1 + US2 fully functional — re-importing an unchanged CSV produces all skips; changing a field produces the correct update count

---

## Phase 5: User Story 3 — Validation Errors (Priority: P3)

**Goal**: Any row-level error (invalid `training_group`, negative hours, missing required column, etc.) is caught before the confirmation prompt; all errors are listed with line numbers in Swedish; no backend changes are made.

**Independent Test**: Create a CSV with (a) an invalid training_group value and (b) negative hours. Run import. Verify CLI lists both errors with correct line numbers and makes no Trailbase changes.

### Implementation for User Story 3

- [ ] T015 [US3] Implement `validateRow(lineNum int, row SessionRow) []ParseError` in `tools/admin-cli/internal/importer/csv.go` covering per-row rules R-01 through R-06 from data-model.md: `month_key` non-empty, `training_group` in allowed set, `date` non-empty, `title` non-empty, `hours` ≥ 0, `minutes` 0–59; return Swedish error messages per data-model.md; **R-07 (required header columns present) is NOT part of this function** — it is a one-time header check in `ParseCSV` that produces a `ParseError{Line: 1}` and causes immediate abort
- [ ] T016 [US3] Extend `ParseCSV` in `tools/admin-cli/internal/importer/csv.go` to call `validateRow` for each data row, accumulate all `ParseError` values across the entire file, and when any errors exist return `(nil, errors, nil)` instead of the row slice (so callers know to abort)
- [ ] T017 [US3] Add `importPhaseError` phase to `importModel` in `tools/admin-cli/internal/ui/import.go`: triggered when `ParseCSV` returns a non-empty error slice; display errors in red with line numbers and the message "Ingen data importerades." before returning to the menu
- [ ] T018 [P] [US3] Unit tests for `validateRow` in `tools/admin-cli/internal/importer/csv_test.go` covering all 7 rules and boundary values: `hours=0` (valid), `minutes=0` and `minutes=59` (valid), `minutes=60` (invalid), each of the 6 allowed `training_group` values (valid), an unknown value (invalid), empty `month_key` / `date` / `title` (each invalid)

**Checkpoint**: All three user stories fully functional — invalid files are rejected with per-row errors; valid files with mixed inserts/updates/skips are applied correctly

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T019 Update `tools/admin-cli/internal/ui/help.go` to include a description of the "Importera tidrapportpass" option, explaining the CSV format and the upsert key
- [ ] T020 [P] Run `go vet ./...` from `tools/admin-cli/` and resolve any reported issues
- [ ] T021 Run `go test ./internal/importer/...` and confirm all unit tests pass with no failures or skips
- [ ] T022 Manual end-to-end validation per `specs/006-import-sessions-csv/quickstart.md`: (a) fresh import → all inserts, (b) re-import unchanged → all skips, (c) change one row → one update, (d) invalid CSV → validation errors displayed, no backend changes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Requires Phase 1 complete — **blocks all user stories**
- **Phase 3 (US1)**: Requires Phase 2 complete — delivers MVP
- **Phase 4 (US2)**: Requires Phase 2 complete — can proceed in parallel with US1 once Phase 2 is done, but US2 extends `ComputeDiff` started in US1 so **sequential is safer**
- **Phase 5 (US3)**: Requires Phase 2 complete — extends `ParseCSV` from US1; sequential recommended
- **Phase 6 (Polish)**: Requires all user story phases complete

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2 — no dependency on other stories
- **US2 (P2)**: Can start after Phase 2 — extends the same `ComputeDiff` and `ApplyImport` functions started in US1; implement after US1 is checkpointed to avoid merge conflicts
- **US3 (P3)**: Can start after Phase 2 — extends `ParseCSV` from US1; implement after US1 is checkpointed

### Within Each Phase

- T001, T002, T003 are [P] and can run together
- T004, T005, T006 in Phase 2: T005 is [P] with T004; T006 depends on T005 (needs `MenuImport` constant)
- In each US phase: implementation tasks before test tasks (tests verify the implementation)

---

## Parallel Execution Examples

### Phase 1 (all parallel)

```
T001: Scaffold csv.go
T002: Scaffold sessions.go    ← run together
T003: Scaffold csv_test.go
```

### Phase 2

```
T004: ListAllSessions (returns []importer.ExistingSession)  ← start first
T005: menu.go changes                                        ← run in parallel with T004
T006: main.go dispatch                                       ← after T005 (needs MenuImport constant)
```

### Phase 3 (US1)

```
T007: ParseCSV happy path     ← start first
T008: ComputeDiff inserts     ← after T007 (uses SessionRow)
T009: CreateSession + Apply   ← after T008 (uses ImportDiff)
T010: TUI importModel         ← after T009 (wires everything together)
T011: Unit tests              ← [P] after T007 (can test ParseCSV independently)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (T007–T011)
4. **STOP and VALIDATE**: Import a fresh CSV → confirm inserts applied
5. Ship MVP — the most common use case (bulk population of a new month) is now covered

### Incremental Delivery

1. Setup + Foundational → project compiles with new menu item
2. US1 → bulk insert works → **MVP shipped**
3. US2 → re-import / correction workflow works
4. US3 → validation errors surfaced cleanly
5. Each story adds value without breaking previous stories

---

## Notes

- [P] marks tasks in different files with no incomplete dependencies — safe to run concurrently
- [US1/2/3] traces each task to its user story for review and rollback scope
- `go build ./...` MUST pass after every phase before moving to the next
- Constitution Principle II: `internal/importer/` is a business-logic package and MUST have unit tests — test tasks T011, T014, T018 are not optional
- All user-visible strings MUST be in Swedish (menu labels, error messages, progress indicators)
- The `minutes` CSV column is optional; default to 0 when absent — this must be explicit in both `ParseCSV` and `validateRow`
