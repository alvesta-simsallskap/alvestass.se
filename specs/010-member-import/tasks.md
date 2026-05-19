# Tasks: Member Register – Data Model and Initial Import

**Input**: Design documents from `specs/010-member-import/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Tests**: Included for all parser and normalizer logic — required by the project constitution for business-logic modules.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no shared dependencies)
- **[Story]**: Which user story this task belongs to (US1/US2/US3)

---

## Phase 1: Setup

**Purpose**: Create the new Go package directory used by all phases.

- [X] T001 Create package directory `tools/admin-cli/internal/memberimporter/` (empty dir with `.gitkeep` or first file)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema, GDPR documentation, and shared Go types must exist before any user story can be tested end-to-end.

**⚠️ CRITICAL**: No user story work can be tested against Trailbase until the migration is applied. The production import is additionally gated on the GDPR register update (T003) and the `/integritetspolicy` page (tracked separately).

- [X] T002 Write Trailbase migration `trailbase/migrations/U1779235200__create_member_register.sql` with all six tables: `members`, `guardians`, `training_groups`, `member_training_groups`, `families`, `family_members` — use the full SQL from `specs/010-member-import/data-model.md` including GDPR migration comments
- [X] T003 Update `docs/gdpr-register.md` with one entry per new table (all six: `members`, `guardians`, `training_groups`, `member_training_groups`, `families`, `family_members`) — personal-data tables (`members`, `guardians`) require full entries covering purpose, data categories, legal basis (Art. 6(1)(b)/(f)), retention, access control, deletion path, and storage location (fly.io arn); relational tables require a brief entry noting no additional personal data fields and referencing the parent table — required GDPR gate before production import (constitution Principle VII; plan GDPR gate requires all six tables)
- [ ] T003b Configure all six new Trailbase tables as authenticated-only via Trailbase Admin (table access settings) — verify that no table is reachable via a public (unauthenticated) Trailbase API route; document the access control setting for each table in `specs/010-member-import/contracts/trailbase-member-tables.md` as ✅ applied — required GDPR gate before production import (constitution Principle VII: "Personal data MUST NOT be served by public unauthenticated API routes")
- [X] T004 [P] Create `tools/admin-cli/internal/memberimporter/model.go` — define all shared intermediate Go types: `RawMember`, `RawGroup`, `RawGuardian`, `ImportPreview`, `MemberImportResult`, `SkippedRecord`; no Trailbase or CSV dependencies
- [X] T005 [P] Create `tools/admin-cli/internal/trailbase/members.go` — implement Trailbase SDK wrappers for upsert/delete of all six new tables using the `tb.NewRecordApi` pattern from `internal/trailbase/sessions.go`; define matching Go structs for each table

**Checkpoint**: Migration written, GDPR register updated, shared types and Trailbase client ready — user story implementation can begin.

---

## Phase 3: User Story 1 – Administrator imports active members (Priority: P1) 🎯 MVP

**Goal**: Parse both CSV exports, identify Deltagare members + board members, deduplicate by IID, write members/guardians/groups to Trailbase, and present a CLI summary.

**Independent Test**: Run `alvestass-admin`, select "Importera memberregister", provide `.temp/Register 260519/` files — verify member count matches the number of unique Deltagare IIDs in the WeUnite export plus board members, and that no Ledare-only rows appear in `members`.

### Tests for User Story 1

- [X] T006 [P] [US1] Write unit tests `tools/admin-cli/internal/memberimporter/idrottonline_test.go` covering: header parsing, member row extraction (IID, name, gender, dob, city, email, phone, member_since), board-role detection (Styrelseledamot/Ordförande/Vice ordförande/Kassör/Sekreterare), skipping of "Till målsman för:" guardian-relationship rows, and rows missing IID
- [X] T007 [P] [US1] Write unit tests `tools/admin-cli/internal/memberimporter/weunite_test.go` covering: Deltagare row collection (Slut date ignored), Ledare/Huvudledare row classification (not Deltagare), guardian slot extraction (up to 3 per row), and semicolon-delimited UTF-8 parsing
- [X] T008 [P] [US1] Write unit tests `tools/admin-cli/internal/memberimporter/normalize_test.go` covering: time-slot stripping regex (various formats: `12.55-13.40`, `13.05–13.50`), no-op for names without time slots, all five category mappings (swim_school/adult/masters/competitive/technique), and unknown group name handling
- [X] T008b [P] [US1] Write unit tests `tools/admin-cli/internal/memberimporter/importer_test.go` covering: (a) deduplication — two Deltagare rows with the same IID produce exactly one `members` record; (b) Ledare-only row (no matching Deltagare) does not appear in the result member list; (c) age filter — member with `date_of_birth` < 18 years produces associated `RawGuardian` records, member ≥ 18 does not; (d) guardian IID resolution — guardian whose personnummer matches a Deltagare in IdrottOnline receives that IID in the result; (e) conflicting IID — same IID in both source files with differing name/email uses IdrottOnline values; (f) board member without WeUnite presence is included; (g) Deltagare with no matching IdrottOnline entry produces a `SkippedRecord` and is excluded from the member list (constitution Principle II: business logic MUST be covered by automated tests)

### Implementation for User Story 1

- [X] T009 [P] [US1] Implement `tools/admin-cli/internal/memberimporter/idrottonline.go` — semicolon-delimited CSV parser: read all rows; skip rows where `Målsman` column is non-empty (guardian-relationship rows); extract IID, first/last name, gender, date of birth (`YYYYMMDD-xxxx` → `YYYY-MM-DD`), city, member_since, email (`E-post kontakt`), phone (`Telefon mobil`), board roles (`Roller` field), and family label (`Familj` field); return `[]RawMember` and `[]SkippedRecord`
- [X] T010 [P] [US1] Implement `tools/admin-cli/internal/memberimporter/weunite.go` — semicolon-delimited CSV parser: collect all rows where `Roll` == `"Deltagare"` regardless of `Slut` date; also collect rows where `Roll` is `"Ledare"` or `"Huvudledare"` for instructor cross-reference; extract personnummer, group name (raw), role, and all three guardian slots (`Målsman 1/2/3` with Förnamn/Efternamn/Telefon/E-post); return `[]RawGroup` (Deltagare) and `[]RawGroup` (instructors)
- [X] T011 [US1] Implement `tools/admin-cli/internal/memberimporter/normalize.go` — (a) strip time-slot suffix from group names using regex `\s+\d{1,2}[.:]\d{2}\s*[-–]\s*\d{1,2}[.:]\d{2}$`; (b) map normalised names to category enum using the table in `research.md §4`; (c) check if a personnummer (WeUnite) matches an IdrottOnline row using a `YYYYMMDD` prefix comparison (first 8 chars of `Födelsedat./Personnr.`)
- [X] T012 [US1] Implement `tools/admin-cli/internal/memberimporter/importer.go` — orchestration: (1) parse both CSVs; (2) build personnummer→RawMember index from IdrottOnline; (3) for each WeUnite Deltagare row look up IID via personnummer — skip with warning if not found; (4) collect board members from IdrottOnline; (5) deduplicate by IID; (6) for each Ledare/Hauptledare row, fetch matching `instructors` entry by email — log warning if none found, skip adding to `members`; (7) collect guardians for members with `date_of_birth` < 18 years; (8) for each guardian extracted from WeUnite, check whether their personnummer exists in the personnummer→RawMember index (guardian may also be a Deltagare) — if found, attach the corresponding IID to the `RawGuardian` record (FR-004); write `role = "participant"` for all `member_training_groups` rows; (9) call Trailbase client upserts; return `MemberImportResult`
- [X] T013 [US1] Implement `tools/admin-cli/internal/ui/importmembers.go` — TUI flow following contract `specs/010-member-import/contracts/cli-import-members.md`: (1) prompt for IdrottOnline CSV path; (2) prompt for WeUnite CSV path; (3) parse both files — on error list all errors and return to menu; (4) show preview (member count, guardian count, group count, skip count); (5) confirm prompt (j/n); (6) on confirm: run import with progress; (7) show result summary with per-category counts and skipped-record list; (8) return to menu
- [X] T014 [US1] Add `MenuImportMembers` constant to `tools/admin-cli/internal/ui/menu.go` and handle it in the main switch in `tools/admin-cli/cmd/alvestass-admin/main.go` calling `ui.RunImportMembers(client)`
- [ ] T015 [US1] Apply migration `U1779235200__create_member_register.sql` to the local/dev Trailbase instance and run end-to-end import test against `.temp/Register 260519/` export files — verify Deltagare count, that no Ledare-only rows appear in `members`, that guardian rows exist for minors, and that group names contain no time slots

**Checkpoint**: User Story 1 fully functional — active members imported, guardians linked, groups normalised, summary shown in CLI.

---

## Phase 4: User Story 2 – System identifies family connections (Priority: P2)

**Goal**: Members sharing the same `Familj` value in IdrottOnline are linked to a shared `families` record, enabling sibling and household identification.

**Independent Test**: After running the full import, query `family_members` in Trailbase — verify that at least one `families` record groups two or more siblings visible in the IdrottOnline export's `Familj` column.

### Implementation for User Story 2

- [X] T016 [US2] Implement `tools/admin-cli/internal/memberimporter/family.go` — group imported members by their `Familj` source label from IdrottOnline; for each non-empty label with ≥ 2 members create one `families` record (storing `source_label`) and one `family_members` row per member; members with empty or unique `Familj` values are skipped
- [X] T017 [US2] Extend `tools/admin-cli/internal/trailbase/members.go` with `UpsertFamily` and `UpsertFamilyMember` methods following the same SDK pattern as other upsert methods
- [X] T018 [US2] Integrate `BuildFamilyLinks()` call into `tools/admin-cli/internal/memberimporter/importer.go` after member upserts complete — add `FamilyLinksImported int` to `MemberImportResult`
- [X] T019 [US2] Update import preview and result summary in `tools/admin-cli/internal/ui/importmembers.go` to display family constellation count in both the preview and the final result
- [ ] T020 [US2] Run end-to-end import and verify family links in Trailbase — cross-check at least one known sibling pair from the IdrottOnline export against `family_members`

**Checkpoint**: User Stories 1 and 2 complete — members imported and family connections established.

---

## Phase 5: User Story 3 – Administrator validates import results (Priority: P3)

**Goal**: The import summary provides enough detail for the administrator to confidently cross-check the result against the source files without manual counting.

**Independent Test**: Run import, review summary — confirm it shows per-role member count (swimmer vs board), per-category group count, and a named list of every skipped record with source file, line number, and reason.

### Implementation for User Story 3

- [X] T021 [US3] Enhance skipped-record reporting in `tools/admin-cli/internal/ui/importmembers.go` — display each `SkippedRecord` with source file name, line number, and Swedish reason string (e.g. "Saknar IID-nummer i IdrottOnline", "Ledare utan matchande post i instruktörstabellen")
- [X] T022 [US3] Add per-role and per-category breakdowns to `MemberImportResult` in `tools/admin-cli/internal/memberimporter/model.go` (e.g. `SwimmersImported int`, `BoardMembersImported int`, `GroupsByCategory map[string]int`) and populate them in `importer.go` and display in `importmembers.go`
- [ ] T023 [US3] Run import and manually verify the summary counts against the source CSV files — swimmer count, board member count, guardian count, group count per category, and skipped-record reasons

**Checkpoint**: All three user stories complete.

---

## Phase N: Polish & Cross-Cutting Concerns

- [X] T024 [P] Run `go test ./...` in `tools/admin-cli/` — all tests must pass
- [X] T025 [P] Run `pnpm build` from repo root — must pass with zero errors (verifies no unintended TypeScript changes)
- [X] T026 Update `specs/010-member-import/research.md` §1 to reflect the final active-member definition: WeUnite Deltagare presence (any group, any date) is the definitive criterion — remove all references to IdrottOnline `Член t.o.m.` as a filter
- [X] T027 Add production-deployment note to `specs/010-member-import/plan.md` — applying the migration and running the import in production requires: (a) `docs/gdpr-register.md` updated (T003), (b) `/integritetspolicy` page published on the site (tracked as `TODO(INTEGRITETSPOLICY)`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — **blocks all user stories**
- **User Stories (Phases 3–5)**: All depend on Phase 2 completion; US1 must precede US2 (family linking extends the importer); US2 must precede US3 (US3 extends the result model)
- **Polish (Phase N)**: Depends on all desired stories being complete

### User Story Dependencies

- **US1 (P1)**: Starts after Phase 2 — no dependency on other stories
- **US2 (P2)**: Starts after US1 — extends `importer.go` and `importmembers.go`
- **US3 (P3)**: Starts after US2 — extends `model.go`, `importer.go`, and `importmembers.go`

### Within Each User Story

- Tests (T006–T008) must be written and verified to **fail** before implementing their target file
- Parsers (T009, T010) before normalizer (T011) before orchestrator (T012)
- Orchestrator (T012) before UI (T013) before menu wiring (T014)
- End-to-end test (T015) last in US1

### Parallel Opportunities

- T004 and T005 can run in parallel (different files, no shared dependencies)
- T003b can run in parallel with T003 (different artefacts — Trailbase admin vs docs)
- T006, T007, T008, T008b (unit tests) can all be written in parallel
- T009, T010 (parsers) can be implemented in parallel once tests are in place

---

## Parallel Example: User Story 1

```
# Write all four test files together:
Task T006:  tools/admin-cli/internal/memberimporter/idrottonline_test.go
Task T007:  tools/admin-cli/internal/memberimporter/weunite_test.go
Task T008:  tools/admin-cli/internal/memberimporter/normalize_test.go
Task T008b: tools/admin-cli/internal/memberimporter/importer_test.go

# Then implement parsers in parallel (after tests fail as expected):
Task T009: tools/admin-cli/internal/memberimporter/idrottonline.go
Task T010: tools/admin-cli/internal/memberimporter/weunite.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002–T005) — apply migration to dev Trailbase
3. Complete Phase 3: US1 (T006–T015)
4. **STOP and VALIDATE**: Run import against `.temp/Register 260519/` — verify member and group counts
5. Address GDPR gate items (T003 + `/integritetspolicy`) before production run

### Incremental Delivery

1. Phase 1 + 2 → Foundation ready (migration applied, types defined)
2. Phase 3 → Import works end-to-end (MVP — members in Trailbase)
3. Phase 4 → Family connections added
4. Phase 5 → Detailed validation report added
5. Polish → Tests green, build passing, production gate documented

---

## Notes

- [P] tasks operate on different files — safe to run concurrently
- The `Slut` date in WeUnite is **ignored** for active-member determination — all Deltagare rows in the export are included
- Personnummer is used in-memory only as a join key between the two CSV files — never written to Trailbase
- "Till målsman för: ..." rows in IdrottOnline are guardian-relationship rows and must be skipped during member parsing
- Production import requires `/integritetspolicy` to be live — this is a separate frontend task not in this feature
- Commit after each phase checkpoint; run `go test ./...` before each commit
