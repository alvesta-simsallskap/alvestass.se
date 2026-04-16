# Tasks: Time Report Trailbase Migration

**Input**: Design documents from `/specs/002-time-report-trailbase/`
**Branch**: `002-time-report-trailbase`

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel with other [P] tasks in the same phase (different files, no shared dependencies)
- **[Story]**: User story this task belongs to (US1/US2/US3)
- All tasks include exact file paths

---

## Phase 1: Setup — Vitest (Constitution Gate II Prerequisite)

**Purpose**: Install the test framework and capture baseline test coverage for the
two business-logic modules BEFORE they are refactored. Constitution Principle II
requires unit tests for all `src/lib/*.ts` before new logic is introduced.

**⚠️ CRITICAL**: No changes to `src/lib/salary.ts` or `src/lib/types.ts` may begin
until this phase is complete.

- [ ] T001 Install `vitest` as a dev dependency (`pnpm add -D vitest`); create `vitest.config.ts` at the repo root with `environment: 'node'` and `include: ['tests/**/*.test.ts']`; add `"test": "vitest run"` to the `scripts` section of `package.json`
- [ ] T002 Create `tests/salary.test.ts` — unit tests for the CURRENT `src/lib/salary.ts` signatures: `findTimeItem`, `buildTable`, `calcSalary` (primary path, boundary values including h=10/15/20 codes, null coachRate, empty checked array); all tests must PASS against the unmodified code
- [ ] T003 [P] Create `tests/timeReportValidation.test.ts` — unit tests for `parseTimeReportForm` in `src/lib/timeReportValidation.ts`: normal form, missing required fields, multiple extra-time rows, empty extra-time rows; all tests must PASS

**Checkpoint**: `pnpm test` exits with 0 failures. Baseline coverage confirmed.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: GDPR gates, database migrations, shared TypeScript types, Trailbase
client extensions, and the salary.ts refactor. ALL of this must be complete before
any user story implementation begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.
The `instructors` table MUST NOT receive production data until T004 and T005 are
committed and `/integritetspolicy` is live.

- [ ] T004 Create `src/pages/integritetspolicy.astro` — a static Astro page that renders a minimal Swedish-language privacy notice (integritetspolicy). The page must be publicly accessible at `/integritetspolicy`, use the existing Bulma layout (import `bulma/css/bulma.min.css` and `_global.scss`), include a heading "Integritetspolicy" and at minimum one paragraph stating that the club processes personal data (instructor salary records) on the legal basis of contractual necessity. Add a link to this page in `src/components/Footer.astro`. **[GDPR gate — must ship before instructors table goes live]**
- [ ] T005 [P] Create `docs/gdpr-register.md` — a register of processing activities (Art. 30 GDPR). Document: (1) `instructors` table: data categories (email, salary rates), legal basis (contractual necessity / employment), retention (end of employment + 1 year), deletion path (Trailbase admin UI), processor (fly.io / Trailbase); (2) Cloudflare Workers: role as data processor, DPA status; (3) Mailjet: role as data processor for email delivery, DPA status. **[GDPR gate — must exist before any personal data is stored]**
- [ ] T006 [P] Create `trailbase/migrations/U1776427200__create_time_report_sessions.sql` — DDL for `time_report_sessions` table. Columns: `id INTEGER PRIMARY KEY`, `month_key TEXT NOT NULL CHECK(month_key != '')`, `training_group TEXT NOT NULL CHECK(training_group IN ('simskola','tavlingA','tavlingB','teknik','masters','vuxencrawl'))`, `date TEXT NOT NULL`, `title TEXT NOT NULL`, `hours INTEGER NOT NULL CHECK(hours >= 0)`, `minutes INTEGER NOT NULL DEFAULT 0 CHECK(minutes >= 0 AND minutes < 60)`. Add `CREATE INDEX idx_sessions_month_group ON time_report_sessions (month_key, training_group)`. Include a GDPR comment: "No personal data — training schedule only". Use `STRICT` table mode.
- [ ] T007 [P] Create `trailbase/migrations/U1776513600__create_instructors.sql` — DDL for `instructors` table. Columns: `id INTEGER PRIMARY KEY`, `email TEXT NOT NULL UNIQUE CHECK(email != '')`, `swim_school_rate INTEGER NOT NULL CHECK(swim_school_rate > 0)`, `coach_rate INTEGER CHECK(coach_rate IS NULL OR coach_rate > 0)`. Include migration comment block: "GDPR: Personal data — email and salary rates. Legal basis: Contractual necessity (employment). Retention: until end of employment + 1 year for accounting. Access: authenticated read (service user); admin-only write. Deletion path: admin deletes row via Trailbase admin UI." Use `STRICT` table mode.
- [ ] T008 [P] Create `trailbase/migrations/U1776600000__create_time_report_config.sql` — DDL for `time_report_config` table. Columns: `id INTEGER PRIMARY KEY`, `active_month_key TEXT NOT NULL`, `active_month_display TEXT NOT NULL`, `extra_time_simskola INTEGER NOT NULL DEFAULT 30`, `extra_time_training INTEGER NOT NULL DEFAULT 15`, `half_day_salary INTEGER NOT NULL DEFAULT 500`, `full_day_salary INTEGER NOT NULL DEFAULT 1000`, `overnight_salary INTEGER NOT NULL DEFAULT 300`. Seed row: `INSERT INTO time_report_config (id, active_month_key, active_month_display) VALUES (1, '2026-04', 'april 2026')`. Include a GDPR comment: "No personal data — operational configuration only." Use `STRICT` table mode.
- [ ] T009 Update `src/lib/types.ts` — (a) rename `Employee` interface to `Instructor`; rename its fields `swimSchoolRate`→`swim_school_rate`, `coachRate`→`coach_rate`; add `id: number` field; (b) add `export type TrainingGroupKey = 'simskola' | 'tavlingA' | 'tavlingB' | 'teknik' | 'masters' | 'vuxencrawl'`; (c) add `export interface Session { date: string; title: string; hours: number; minutes: number; }`; (d) add `export type SessionSchedule = Record<TrainingGroupKey, Session[]>`; (e) add `export interface TimeReportConfig { id: number; active_month_key: string; active_month_display: string; extra_time_simskola: number; extra_time_training: number; half_day_salary: number; full_day_salary: number; overnight_salary: number; }`. Remove unused `Employee` export.
- [ ] T010 Refactor `src/lib/salary.ts` — (a) remove all top-level imports of `time-report-items.json` and `time-report-settings.ts`; (b) update `findTimeItem(schedule: SessionSchedule, group: TrainingGroupKey, value: string): Session | undefined` (was `findTimeItem(section, value)`, now pure — no module-level data); (c) update `buildTable(group: TrainingGroupKey, label: string, checked: string[], schedule: SessionSchedule, config: TimeReportConfig)` to accept schedule and config params instead of reading globals; (d) update `calcSalary(group: TrainingGroupKey | 'extratid', checked: string[], schedule: SessionSchedule, config: TimeReportConfig, instructor?: Instructor, extraRows?: ExtraTimeRow[])` — replace `employee.swimSchoolRate` with `instructor.swim_school_rate`, `employee.coachRate` with `instructor.coach_rate`; replace `EXTRA_TIME_SIMSKOLA`/`EXTRA_TIME_TRAINING` with `config.extra_time_simskola`/`config.extra_time_training`; eliminate all remaining `any` usages; import `TrainingGroupKey`, `Session`, `SessionSchedule`, `TimeReportConfig`, `Instructor` from `./types`
- [ ] T011 Update `tests/salary.test.ts` to match the new function signatures introduced in T010. Tests must continue to PASS. Construct minimal `SessionSchedule` and `TimeReportConfig` fixtures as test helpers. Keep the same scenario coverage (primary path, boundary values h=10/15/20, null coach_rate, empty checked).
- [ ] T012 [P] Extend `src/lib/trailbase.ts` — add four new exported async functions below the existing `fetchClubInfo`: (1) `authenticateServiceUser(baseUrl: string, email: string, password: string): Promise<string>` — POSTs to `/api/auth/v1/token` with `{email, password}`, returns `auth_token` from response; throws on non-2xx; (2) `fetchTimeReportConfig(baseUrl: string, authToken: string): Promise<TimeReportConfig | null>` — GETs `/api/records/v1/time_report_config?limit=1` with Bearer token, returns first record or null; (3) `fetchTimeReportSessions(baseUrl: string, monthKey: string, authToken: string): Promise<Session[]>` — GETs `/api/records/v1/time_report_sessions?filters=month_key%3D%3D{monthKey}&limit=500` with Bearer token, returns records array (empty array on 0 results); maps Trailbase response fields (`hours`, `minutes`) to `Session` shape; (4) `fetchInstructor(baseUrl: string, email: string, authToken: string): Promise<Instructor | null>` — GETs `/api/records/v1/instructors?filters=email%3D%3D{encodeURIComponent(email)}&limit=1` with Bearer token, returns first record or null. Import `TimeReportConfig`, `Session`, `Instructor` from `./types`. All functions must throw (not return null) on network errors; callers handle null for "not found" and catch for errors.
- [ ] T013 [P] Update `src/env.d.ts` — add `TRAILBASE_SERVICE_EMAIL: string` and `TRAILBASE_SERVICE_PASSWORD: string` to the `env` block inside the `App.Locals` declaration, alongside the existing secrets

**Checkpoint**: `pnpm test` still passes (T011 updated tests green). `pnpm build` passes (no TS errors from the types/lib changes). Types and Trailbase client are ready for the page changes.

---

## Phase 3: User Story 1 — Admin Publishes Monthly Schedule (Priority: P1) 🎯 MVP

**Goal**: Replace the static `time-report-items.json` data source with live sessions
from the Trailbase `time_report_sessions` table. The form loads its schedule on
every request; no deployment is required to update sessions.

**Independent Test** (from spec.md): Create a new session in the Trailbase admin UI
for the active month, reload `/tidrapport`, confirm the new session appears in the
correct group with the correct date and title.

- [ ] T014 [US1] Add `export const prerender = false;` as the very first line of `src/pages/tidrapport.astro`, directly followed by a comment: `// SSR required: schedule and config are fetched from Trailbase at request time (FR-001, FR-002)`
- [ ] T015 [US1] Add an Astro frontmatter block to `src/pages/tidrapport.astro` (after the `prerender` export) that: (a) reads `TRAILBASE_URL`, `TRAILBASE_SERVICE_EMAIL`, `TRAILBASE_SERVICE_PASSWORD` from `Astro.locals.runtime.env`; (b) calls `authenticateServiceUser`, then `Promise.all([fetchTimeReportConfig, fetchTimeReportSessions])` using the auth token; (c) catches any thrown error or null config and sets `const error = true`; (d) groups the returned `Session[]` into a `Record<TrainingGroupKey, Session[]>` object by `training_group`; (e) maps each group's sessions into the shape the `TimeReportCheckboxGroup` component expects: `Array<{ date: string; title: string; h: number; m: number }>` (using `hours` and `minutes` from `Session`). Import `authenticateServiceUser`, `fetchTimeReportConfig`, `fetchTimeReportSessions` from `../../lib/trailbase` and the relevant types from `../../lib/types`.
- [ ] T016 [US1] Add a `Astro.response.headers.set('Cache-Control', 'no-store')` call in the frontmatter of `src/pages/tidrapport.astro` (required per constitution Gate IV — schedule must be fresh on every load). Add a Swedish error `<div class="notification is-danger">` rendered when `error === true` that reads: "Det gick inte att hämta schemat just nu. Försök igen om en stund." Wrap the existing `<form>` in `{!error && (...)}` so it is hidden when the backend is unreachable.
- [ ] T017 [US1] Remove the static imports from `src/pages/tidrapport.astro`: delete `import timeReportItems from '../config/time-report-items.json'` and `import { TIME_REPORT_MONTH_KEY, TIME_REPORT_MONTH_DISPLAY, IS_DEVELOPMENT } from '../config/time-report-settings'`. Replace `const items = timeReportItems[TIME_REPORT_MONTH_KEY]` with the grouped sessions from the frontmatter. Replace `IS_DEVELOPMENT` references with `import.meta.env.DEV`. Replace `TIME_REPORT_MONTH_DISPLAY` in `<title>` and `<h2>` with `config?.active_month_display ?? ''`.
- [ ] T018 [US1] Update each `<TimeReportCheckboxGroup>` call in `src/pages/tidrapport.astro` to use the mapped session arrays from the frontmatter (e.g., `items={mappedSessions.simskola}`) instead of `items.simskola` etc. Ensure the `section` prop values remain unchanged (they drive form field names used in `timeReportValidation.ts`).

**Checkpoint**: `pnpm dev`, open `http://localhost:4321/tidrapport`. Form loads with schedule data from Trailbase. Error notice appears when Trailbase is unreachable. `pnpm build` passes with zero errors.

---

## Phase 4: User Story 2 — Admin Updates Instructor Salary Rates (Priority: P2)

**Goal**: Replace the hardcoded `EMPLOYEES` array in `send-time-report.ts` with a
live lookup against the Trailbase `instructors` table. Rate changes by the admin
are reflected in the next submitted time report without a deployment.

**Independent Test** (from spec.md): Update one instructor's `swim_school_rate` in
the Trailbase admin UI. Submit a time report with that instructor's email. Verify
the preliminary salary estimate in the received email uses the updated rate.

- [ ] T019 [US2] Add service user authentication to `src/pages/api/send-time-report.ts`: at the top of the `POST` handler, read `TRAILBASE_URL`, `TRAILBASE_SERVICE_EMAIL`, `TRAILBASE_SERVICE_PASSWORD` from `locals.runtime.env`; call `authenticateServiceUser(TRAILBASE_URL, email, password)`; wrap in try/catch so auth failure does NOT abort the submission (salary estimate is omitted on auth failure per spec). Replace `IS_DEVELOPMENT` import (removed from `time-report-settings.ts`) with `import.meta.env.DEV`. Import `authenticateServiceUser`, `fetchInstructor`, `fetchTimeReportConfig` from `../../lib/trailbase`.
- [ ] T020 [US2] Replace the hardcoded `EMPLOYEES` array in `src/pages/api/send-time-report.ts` with a call to `fetchInstructor(TRAILBASE_URL, data.email, authToken)`. If auth succeeded, attempt the lookup; if lookup throws or returns null, proceed without instructor (salary estimate omitted). Delete the entire `EMPLOYEES` array literal.
- [ ] T021 [US2] Fetch `TimeReportConfig` in `src/pages/api/send-time-report.ts` using `fetchTimeReportConfig(TRAILBASE_URL, authToken)`. Use `config.active_month_display` for the email HTML heading (`<h4>Tidrapport ${config.active_month_display}</h4>`) and `config.active_month_key` for the email subject in `sendTimeReportEmail` (replace `TIME_REPORT_MONTH_KEY` and `TIME_REPORT_MONTH_DISPLAY` imports). If config fetch fails, fall back to empty string month display and omit salary estimate.
- [ ] T022 [US2] Fetch the current month's sessions in `src/pages/api/send-time-report.ts` by calling `fetchTimeReportSessions(TRAILBASE_URL, config.active_month_key, authToken)` (the Worker already has `authToken` and `config` from T019/T021). Group the returned `Session[]` by `training_group` into a `SessionSchedule` (same grouping logic as T015 in `tidrapport.astro`). Then update all `buildTable`, `calcSalary`, and `findTimeItem` calls to pass this `schedule` and `config` as parameters (matching the refactored signatures from T010). Replace all `HALF_DAY_SALARY`, `FULL_DAY_SALARY`, `OVERNIGHT_SALARY` references with `config.half_day_salary`, `config.full_day_salary`, `config.overnight_salary`. Remove any remaining imports from `time-report-settings.ts`. Note: the form POST only contains checked date+title strings — hours and minutes are NOT submitted by the browser, so a server-side re-fetch is required for salary calculation.

**Checkpoint**: Submit a time report locally (`pnpm dev`, IS_DEVELOPMENT mode). HTML preview shows correct salary estimate. No hardcoded salary or rate data remains in source files.

---

## Phase 5: User Story 3 — Admin Changes Active Reporting Period (Priority: P3)

**Goal**: Verify that changing `active_month_key` / `active_month_display` in the
Trailbase config immediately updates both the form heading and the email — with
no code change or redeployment.

**Independent Test** (from spec.md): Update `active_month_key` and
`active_month_display` in `time_report_config` via the admin UI. Reload
`/tidrapport`. Confirm the `<title>`, `<h2>`, and loaded sessions all reflect
the new month.

- [ ] T023 [US3] Audit `src/pages/tidrapport.astro` — confirm `<title>` and `<h2 class="subtitle">` both use `config?.active_month_display ?? ''` (not any constant from `time-report-settings.ts`). If either still references a hardcoded value, fix it. Confirm the sessions fetch passes `config.active_month_key` (not a hardcoded string) to `fetchTimeReportSessions`.
- [ ] T024 [US3] Audit `src/pages/api/send-time-report.ts` — confirm both the email subject line (`Subject: \`Tidrapport för ${data.name} ${config.active_month_key}\``) and email heading use the dynamically fetched config values (no remaining references to `TIME_REPORT_MONTH_KEY` or `TIME_REPORT_MONTH_DISPLAY`). If any remain, fix them.

**Checkpoint**: Change `active_month_key` to `'2026-05'` and `active_month_display` to `'maj 2026'` in Trailbase admin. Reload `/tidrapport` — heading reads "maj 2026", session list is empty (no sessions for 2026-05). Reset config to `2026-04` before submitting any real reports.

---

## Polish & Cross-Cutting Concerns

**Purpose**: Remove dead code, validate the full build, and run end-to-end acceptance.

- [ ] T025 Delete `src/config/time-report-items.json` — only after confirming all sessions for the current active month have been entered in the Trailbase `time_report_sessions` table (verify via Trailbase admin UI)
- [ ] T026 [P] Delete `src/config/time-report-settings.ts` — only after confirming no remaining imports of this file exist anywhere in `src/` (run `grep -r "time-report-settings" src/` first)
- [ ] T027 Run `pnpm build` — fix all TypeScript strict-mode errors and `astro check` warnings until the command exits cleanly with zero errors. This is the merge gate (SC-005).
- [ ] T028 Manual end-to-end test per `specs/002-time-report-trailbase/quickstart.md` steps 7–9: (a) `curl` config and sessions endpoints to verify authenticated access; (b) `pnpm dev`, load `/tidrapport`, confirm schedule renders; (c) submit form with a registered instructor email in development mode, verify HTML preview shows correct salary estimate (SC-006 email flow regression check).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 completion (T001–T003 done)
  - **Blocks all user stories**
  - T009 must complete before T010 and T012 (type definitions needed)
  - T010 must complete before T011 (tests reflect new signatures)
- **Phase 3 (US1)**: Depends on Phase 2 completion
- **Phase 4 (US2)**: Depends on Phase 2 completion; can start in parallel with Phase 3
- **Phase 5 (US3)**: Depends on Phase 3 and Phase 4 (audits confirm their output)
- **Polish**: Depends on Phases 3–5 completion

### User Story Dependencies

- **US1 (T014–T018)**: Can begin as soon as Phase 2 is done — no US2/US3 dependency
- **US2 (T019–T022)**: Can begin as soon as Phase 2 is done — no US1 dependency (different files)
- **US3 (T023–T024)**: Depends on US1 and US2 (audits their output); minimal standalone work

### Within Phase 2

- T004 and T005 [P]: can run together (different files)
- T006, T007, T008 [P]: can run together (different migration files)
- T009 → T010 → T011: sequential (types before salary.ts before updating tests)
- T012 [P] with T006–T008: can run in parallel (different files)
- T013 [P] with T012: can run in parallel

### Parallel Opportunities

- T002 ‖ T003: different test files
- T004 ‖ T005 ‖ T006 ‖ T007 ‖ T008 ‖ T012 ‖ T013: all different files (once T009 is done)
- T014–T018 ‖ T019–T022: different page files (US1 and US2 in parallel)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (Vitest setup)
2. Complete Phase 2 (Foundational — critical gate)
3. Complete Phase 3 (US1 — live schedule)
4. **STOP and VALIDATE**: Confirm schedule loads from Trailbase; `pnpm build` green
5. US1 is shippable independently — schedule updates no longer need deployment

### Incremental Delivery

1. Setup + Foundational → types, migrations, and Trailbase client ready
2. US1 → live schedule on form (deployable increment)
3. US2 → live salary rates in email (deployable increment, removes personal data from source)
4. US3 → active month is dynamic (no new code — verified in audit tasks)
5. Polish → dead code removed, build gate confirmed

---

## Notes

- `[P]` tasks touch different files and have no in-flight dependencies — safe to run in parallel
- GDPR gates (T004, T005) are code-level prerequisites: the `instructors` table must not receive production data until `/integritetspolicy` is live and `docs/gdpr-register.md` is committed
- The `TimeReportCheckboxGroup` component interface (`h`, `m`) is NOT changed — mapping from `Session.hours`/`Session.minutes` to `h`/`m` happens in the `tidrapport.astro` frontmatter (T015)
- Commit after each phase checkpoint to maintain a recoverable history
