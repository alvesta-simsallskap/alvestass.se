# Tasks: Fix Övrig Tid Minutes Bug

**Input**: Design documents from `/specs/008-fix-ovrig-tid-minutes/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅

**Organization**: Two user stories. US1 (core fix) must be done first; US2 verifies the empty-row guard is preserved. Both changes are in two files and can be done in a single sitting.

## Format: `[ID] [P?] [Story] Description`

---

## Phase 1: Setup

**Purpose**: No new project structure needed. This phase confirms the working environment before touching code.

- [x] T001 Confirm `pnpm build` passes on `008-fix-ovrig-tid-minutes` before any changes

---

## Phase 2: User Story 1 — Report whole-hour extra time (Priority: P1) 🎯 MVP

**Goal**: An Övrig tid row submitted with an empty minutes (or hours) field is no longer silently dropped; it is included in the report with the blank field treated as 0.

**Independent Test**: Submit the time-report form with one Övrig tid row (valid date, 2 h, empty minutes, a description). Verify the resulting email shows "2 h 0 m" and the salary calculation reflects 2 hours.

### Implementation

- [x] T002 [US1] Fix validity check and coerce empty h/m to "0" in `src/lib/timeReportValidation.ts`  
  — Change `if (date && h && m && desc)` to `if (date && desc)` and coerce `h`/`m` empty strings to `"0"` before building the row.

- [x] T003 [P] [US1] Default new extraTimes rows to `{ h: 0, m: 0 }` instead of `{ h: '', m: '' }` in `src/pages/tidrapport.astro`  
  — Update both the initial row (line ~66) and the "Lägg till rad" push (line ~256) so Alpine.js initialises numeric fields with `0`.

**Checkpoint**: T002 and T003 are independent (different files) and can be done in parallel. After both: run `pnpm build`, then do a manual end-to-end test with empty minutes in the `pnpm dev` debug mode.

---

## Phase 3: User Story 2 — Partial rows excluded (Priority: P2)

**Goal**: Fully empty Övrig tid rows and rows missing a date or description are still excluded from the submitted report after the Phase 2 changes.

**Independent Test**: Submit the form with one completely blank Övrig tid row and one properly filled row. Verify only the filled row appears in the email.

### Implementation

- [x] T004 [US2] Review `src/lib/timeReportValidation.ts` after T002 — confirm the validity check still requires `date` and `desc` to be non-empty, and that a row with both fields blank is excluded  
  — No code change expected; this is a reading/verification task to confirm the fix doesn't regress the existing guard.

**Checkpoint**: US2 is satisfied by design of the T002 change. Verify by inspection and via the manual e2e test.

---

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T005 [P] Run `pnpm build` and confirm zero TypeScript errors and zero `astro check` warnings after all changes
- [x] T006 Manual end-to-end browser test: submit a time report with (a) empty minutes, (b) empty hours, (c) both set to 0, (d) a fully blank row — verify correct email output for each case using `pnpm dev` debug mode
- [x] T007 [P] Verify the Övrig tid UI at mobile (≤768 px), tablet, and desktop — confirm number inputs with default `0` render correctly at all breakpoints

---

## Dependencies & Execution Order

- **T001**: Start immediately — no dependencies
- **T002, T003**: Can run in parallel after T001 (different files)
- **T004**: After T002 (reviewing the same file)
- **T005, T006, T007**: After T002 + T003

### Parallel Opportunities

```
T001
 ├── T002  (timeReportValidation.ts)
 └── T003  (tidrapport.astro)
      └── T004  (review timeReportValidation.ts)
           ├── T005  [P] pnpm build
           ├── T006  e2e manual test
           └── T007  [P] responsive UI check
```

---

## Implementation Strategy

### MVP (User Story 1 only)

1. T001 — confirm clean build baseline
2. T002 + T003 (in parallel) — fix validation + fix UI default
3. T005 — confirm build still passes
4. T006 — manual e2e test

This delivers a fully working fix. T004 and T007 add confidence but are not blocking.

### Full delivery

Complete all tasks T001–T007 before opening the PR.

---

## Notes

- No new files or directories are created by this fix
- No Trailbase migrations, no schema changes, no new dependencies
- Constitution Principle II requires a manual e2e test of the complete form submission (T006) before merge — this is non-negotiable
- TODO(TEST_FRAMEWORK): unit tests for `timeReportValidation.ts` are deferred; the fix must remain simple enough to verify by inspection
