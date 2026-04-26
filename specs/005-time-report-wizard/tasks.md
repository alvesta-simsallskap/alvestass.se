# Tasks: Two-Step Time Report Wizard

**Input**: Design documents from `specs/005-time-report-wizard/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Organization**: Tasks are grouped by user story to enable independent
implementation and testing of each increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to
- Exact file paths are included in every task description

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Schema migration and TypeScript type updates that ALL user story
phases depend on. No user story work can start until T001–T003 are complete.

**⚠️ CRITICAL**: Run `pnpm build` after T002 and T003 to catch type errors early.

- [ ] T001 Write Trailbase migration `trailbase/migrations/U1776686400__update_instructors.sql`: recreate `instructors` table with `swim_school_rate` nullable (remove NOT NULL), add `travel_compensation INTEGER NOT NULL DEFAULT 0`, add CHECK constraint `(swim_school_rate IS NOT NULL OR coach_rate IS NOT NULL)` — use table-rename recreation pattern per research.md Decision 1
- [ ] T002 [P] Update `src/lib/types.ts`: change `Instructor.swim_school_rate` from `number` to `number | null`; add `travel_compensation: boolean` field to the `Instructor` interface
- [ ] T003 [P] Update `src/lib/salary.ts`: in `calcSalary`, the line `if (group === 'simskola') rate = instructor.swim_school_rate` already works with `number | null` via the existing `rate ? ... : 0` guard — verify TypeScript strict mode accepts the updated `Instructor` type with no additional changes needed

**Checkpoint**: Foundation ready — all user story phases can now begin

---

## Phase 2: Step 1 Email Lookup (Prerequisite for All User Stories)

**Purpose**: Implements the step 1 email lookup endpoint and the Alpine.js
state machine. This phase is the gateway to all story phases — it enables
step 2 to render.

- [ ] T004 Create `src/pages/api/lookup-instructor.ts`: SSR API route (`export const prerender = false`) that accepts `POST { email: string }`, authenticates via `authenticateServiceUser`, calls `fetchInstructor` from `src/lib/trailbase.ts`, and returns `{ swimSchool: boolean, coach: boolean, travelCompensation: boolean }` (200) or `{ error: "not_found" }` (404) or `{ error: "backend_unavailable" }` (503) or `{ error: "invalid_email" }` (400) — MUST NOT log the email value; set `Cache-Control: no-store`; see `contracts/lookup-instructor.md` for full contract
- [ ] T005 Rewrite `src/pages/tidrapport.astro`: replace the existing single-step form with a top-level Alpine.js `x-data` component holding state `{ step: 1, email: '', loading: false, error: null, role: null, extraTimes: [...], files: [...] }` and a `lookupEmail()` async method that POSTs to `/api/lookup-instructor`, sets `this.role` on success and advances `this.step` to 2, or sets `this.error` (Swedish message) on 404/503 — step 1 UI: email `<input>` bound to `x-model="email"`, submit button disabled (`x-bind:disabled="loading"`) with loading text `x-text="loading ? 'Söker...' : 'Fortsätt'"`, and error notification `<div x-show="error" x-text="error">` — do NOT add step 2 content yet; verify step 1 submits and shows error for unknown email

---

## Phase 3: User Story 1 — Swim School Leader Form (Priority: P1) 🎯 MVP

**Goal**: After entering a registered swim-school-only email, the instructor
sees Simskola + Övrig tid + Kommentarer and can submit a complete time report.

**Independent Test**: Register an instructor with `swim_school_rate` set and
`coach_rate` NULL in Trailbase. Enter their email in step 1. Confirm only
Simskola, Övrig tid, and Kommentarer appear in step 2. Submit; verify the
email is received at the payroll inbox.

- [ ] T006 [US1] Add step 2 form base to `src/pages/tidrapport.astro` (visible when `step === 2`): include name field (`name="namn"`), hidden email field (`:value="email"` `name="email"` `type="hidden"`), Övrig tid section (always — keep existing Alpine state `extraTimes` merged into top-level `x-data` component, adapt `collectExtraTime` and `handleSubmit` methods into the same top-level component), Kommentarer textarea (`name="kommentarer"`, always visible), Turnstile widget, and submit button — replaces the existing `formHandler` Alpine component
- [ ] T007 [US1] Add Simskola section to step 2 in `src/pages/tidrapport.astro` using `<template x-if="role?.swimSchool">` wrapping the existing Simskola box markup (with `TimeReportCheckboxGroup` component, `mappedSessions.simskola` items, and the preparation-time note)
- [ ] T008 [US1] Browser verify using `pnpm dev`: swim-school-only instructor (coach_rate NULL in Trailbase) — step 1 accepts email → step 2 shows Simskola and Övrig tid and Kommentarer → Träningsgrupper and Utlägg are absent from the page → submit opens email preview in new tab (dev mode)

---

## Phase 4: User Story 2 — Coach Form (Priority: P2)

**Goal**: After entering a coach-only email, the instructor sees Träningsgrupper
+ Utlägg + Övrig tid + Kommentarer (no Simskola).

**Independent Test**: Register an instructor with `coach_rate` set and
`swim_school_rate` NULL. Enter their email in step 1. Confirm Träningsgrupper
and Utlägg appear; Simskola is absent. Submit; verify the email is received.

- [ ] T009 [US2] Add Träningsgrupper section to step 2 in `src/pages/tidrapport.astro` using `<template x-if="role?.coach">` wrapping the existing Träningsgrupper box markup (all training group columns: tavlingA, tavlingB, teknik, masters, vuxencrawl, plus preparation-time note)
- [ ] T010 [US2] Add Utlägg section to step 2 in `src/pages/tidrapport.astro` using `<template x-if="role?.coach">` wrapping the existing Utlägg box markup (file upload inputs with Alpine `files` state — merge `files` array into top-level `x-data` component)
- [ ] T011 [US2] Browser verify using `pnpm dev`: coach-only instructor (swim_school_rate NULL in Trailbase) — step 2 shows Träningsgrupper + Utlägg + Övrig tid + Kommentarer; Simskola is absent from the page; submit works

---

## Phase 5: User Story 4 — Travel Compensation (Priority: P3)

**Goal**: The Milersättning field appears only when `role.travelCompensation`
is true; it is absent from the DOM (not CSS-hidden) for ineligible instructors.

**Independent Test**: Two instructors — one with `travel_compensation = 1`,
one with `travel_compensation = 0`. Confirm field visible for eligible
instructor (inspect DOM) and absent from DOM for ineligible instructor.

- [ ] T012 [US4] Add Milersättning field to step 2 in `src/pages/tidrapport.astro` using `<template x-if="role?.travelCompensation">` wrapping the Milersättning input (`name="milersattning"`) and its label/help text — place it in the same columns box as Kommentarer (existing layout), using `x-if` to ensure DOM removal for ineligible instructors (satisfies SC-003)
- [ ] T013 [US4] Browser verify: eligible instructor sees Milersättning; browser DevTools confirm the field is absent from the DOM for ineligible instructor (not just display:none)

---

## Phase 6: User Story 3 — Both Roles Verification (Priority: P3)

**Goal**: An instructor with both rates set sees all sections: Simskola +
Träningsgrupper + Övrig tid + Utlägg + Kommentarer.

**Independent Test**: Register an instructor with both `swim_school_rate` and
`coach_rate` set. Verify all sections appear in step 2 and a full submission
(sessions from both sections) produces a correct email.

- [ ] T014 [US3] Browser verify using `pnpm dev`: both-role instructor — step 2 shows Simskola + Träningsgrupper + Övrig tid + Utlägg + Kommentarer; check sessions from both sections + extra time entry → submit → email preview shows all reported groups (no code changes expected; this is a validation task)

---

## Phase 7: User Story 5 — Error Handling Verification (Priority: P2)

**Goal**: Unknown email shows a Swedish error message and keeps the user on
step 1. Backend unavailability also shows a Swedish error.

**Independent Test**: Enter an unregistered email; confirm error message
appears and step 2 is not shown. Verify message is in Swedish.

- [ ] T015 [US5] Browser verify using `pnpm dev`: enter an email not in Trailbase → Swedish error message shown → step 2 not rendered; correct the email → form advances; verify the loading state (button disabled, text changes to "Söker...") is visible during the lookup (may need to add artificial latency via browser DevTools throttling)

---

## Phase 8: Polish & Cross-Cutting Concerns

- [ ] T016 Run `pnpm build` (runs `wrangler types && astro check && astro build`) and resolve any TypeScript strict-mode errors from the `Instructor` type changes — verify zero errors and zero `astro check` warnings
- [ ] T017 [P] Verify no email addresses appear in Cloudflare Worker log output from `lookup-instructor.ts` — review the implementation for any accidental `console.log` / `console.error` calls that include the email value
- [ ] T018 [P] Add `travel_compensation` field to GDPR notes in `trailbase/migrations/U1776686400__update_instructors.sql` migration comment (legal basis: contractual necessity; retention: end of employment + 1 year) — confirm TODO(GDPR_REGISTER) note is present

---

## Dependency Graph

```
T001 ─────────────────────────────────────┐
T002 [P] ─────────────────────────────────┤
T003 [P] ─────────────────────────────────┤
                                           ▼
                                     T004, T005  (Step 1)
                                           │
                    ┌──────────────────────┤
                    ▼                      ▼
             T006, T007, T008       T009, T010, T011
               (US1 — P1)              (US2 — P2)
                    │                      │
                    └──────────┬───────────┘
                               ▼
                         T012, T013     T014       T015
                          (US4 — P3)  (US3 — P3)  (US5 — P2)
                               │
                               ▼
                         T016, T017, T018 (Polish)
```

---

## Parallel Execution Opportunities

**Phase 1** (after T001 completes): T002 and T003 can run simultaneously  
**Phase 3+4** (after T004+T005): US1 and US2 are both in `tidrapport.astro`
so they must be sequential within the file, but can be done in one sitting  
**Polish**: T017 and T018 are independent of each other

---

## Implementation Strategy

**MVP (User Story 1)**: Complete T001–T008. This delivers the full two-step
flow for the most common role (swim school leaders) including the email lookup,
error handling, and correct section visibility.

**Full delivery**: Add T009–T015 to cover coaches, travel compensation, and
both-role instructors.

**Total tasks**: 18  
**Tasks per story**: US1: 3 | US2: 3 | US3: 1 | US4: 2 | US5: 1 | Foundational: 5 | Polish: 3
