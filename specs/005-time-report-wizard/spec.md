# Feature Specification: Two-Step Time Report Wizard

**Feature Branch**: `005-time-report-wizard`
**Created**: 2026-04-26
**Status**: Draft
**Input**: User description: "The time report should be a two step process. First, verify the email adress, then show the relevant time report input fields based on if it is only a swim school leader, only a coach, or both. Travel compensation ("milersättning") should only be visible for instructors who has a special attribute "travel_compensation" set to true."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Swim School Leader Submits Time Report (Priority: P1)

An instructor who only leads swim school sessions opens the time report page.
They enter their registered email address in the first step. The system
identifies them as a swim school leader and advances to the second step, which
shows only the Simskola section — no training group sections are visible.
The instructor fills in their sessions and submits the report.

**Why this priority**: This is the most common instructor role in the club.
Hiding irrelevant training group sections reduces confusion and entry errors.
It also validates the entire two-step flow end-to-end.

**Independent Test**: Register an instructor with swim school duties only in
the backend. Open the time report form, enter that instructor's email, and
confirm that only the Simskola section appears in step 2. Submit the form and
confirm the email is received.

**Acceptance Scenarios**:

1. **Given** an instructor is registered with swim school duties only
   (`swim_school_rate` set, `coach_rate` NULL),
   **When** they enter their email in step 1 and proceed,
   **Then** the form shows the Simskola and Övrig tid sections; no
   Träningsgrupper or Utlägg section is visible.

2. **Given** the instructor fills in one or more swim school sessions and
   submits,
   **When** the submission is processed,
   **Then** a time report email is delivered to the payroll inbox.

3. **Given** the instructor does not have `travel_compensation` set to true,
   **When** the form is shown,
   **Then** the Milersättning field is not visible.

---

### User Story 2 — Coach Submits Time Report (Priority: P2)

An instructor who only coaches training groups opens the time report page.
They enter their registered email in step 1. The system identifies them as a
coach and shows only the Träningsgrupper section in step 2 — no Simskola
section is visible.

**Why this priority**: Ensures coaches see only their relevant sections,
reducing form length and eliminating the risk of accidentally reporting swim
school hours they did not work.

**Independent Test**: Register an instructor with coaching duties only in the
backend. Open the form, enter their email, and confirm only the Träningsgrupper
section appears in step 2. Submit and confirm the email is received.

**Acceptance Scenarios**:

1. **Given** an instructor is registered with coaching duties only
   (`coach_rate` set, `swim_school_rate` NULL),
   **When** they enter their email in step 1 and proceed,
   **Then** the form shows the Träningsgrupper, Övrig tid, and Utlägg sections;
   no Simskola section is visible.

2. **Given** the instructor fills in sessions from one or more training groups
   and submits,
   **When** the submission is processed,
   **Then** a time report email is delivered to the payroll inbox.

---

### User Story 3 — Instructor With Both Roles Submits Time Report (Priority: P3)

An instructor who both leads swim school and coaches training groups enters
their email in step 1. The system identifies them as holding both roles and
shows both the Simskola section and the Träningsgrupper section in step 2.

**Why this priority**: Some instructors hold both roles. The form must handle
this combination without requiring a separate workflow.

**Independent Test**: Register an instructor with both swim school and coaching
duties in the backend. Open the form, enter their email, and confirm both
sections appear in step 2.

**Acceptance Scenarios**:

1. **Given** an instructor is registered with both swim school and coaching
   duties (both `swim_school_rate` and `coach_rate` set),
   **When** they enter their email in step 1 and proceed,
   **Then** the form shows the Simskola, Träningsgrupper, Övrig tid, and
   Utlägg sections.

2. **Given** the instructor reports hours in both sections and submits,
   **When** the submission is processed,
   **Then** a time report email covering both roles is delivered to the payroll
   inbox.

---

### User Story 4 — Eligible Instructor Claims Travel Compensation (Priority: P3)

An instructor with `travel_compensation = true` on their profile opens the
form, completes step 1, and is shown the Milersättning (travel compensation)
field alongside the rest of the form in step 2.

An instructor without this attribute does not see the field at all.

**Why this priority**: Travel compensation is a contractual entitlement for a
subset of instructors. Showing it only to eligible individuals prevents
incorrect claims while ensuring eligible instructors can always claim it.

**Independent Test**: Register two instructors — one with `travel_compensation
= true` and one without. Open the form for each and confirm the field appears
only for the eligible instructor.

**Acceptance Scenarios**:

1. **Given** an instructor's profile has `travel_compensation = true`,
   **When** the form is shown in step 2,
   **Then** the Milersättning field is visible and submittable.

2. **Given** an instructor's profile does not have `travel_compensation = true`,
   **When** the form is shown in step 2,
   **Then** the Milersättning field is not visible.

3. **Given** an eligible instructor enters a travel compensation amount and
   submits,
   **When** the submission is processed,
   **Then** the submitted amount is included in the time report email.

---

### User Story 5 — Unknown Email Is Handled Gracefully (Priority: P2)

An instructor enters an email address that is not registered in the backend.
The system shows a clear Swedish error message in step 1 and allows them to
correct the address. They do not proceed to step 2 until a registered email
is provided.

**Why this priority**: Entering the wrong email is a likely mistake. The error
path must be clear so the instructor can self-correct without contacting anyone.

**Independent Test**: Enter an email that does not exist in the backend and
confirm the error message appears and step 2 is not shown.

**Acceptance Scenarios**:

1. **Given** an email address is not registered in the backend,
   **When** the instructor submits it in step 1,
   **Then** a clear Swedish error message is displayed and the form stays on
   step 1.

2. **Given** the error is shown,
   **When** the instructor corrects the email and resubmits step 1,
   **Then** if the new email is found, the form advances to step 2.

---

### Edge Cases

- What happens if the backend is unreachable during the email lookup in step 1?
  The system shows a clear Swedish error message and the instructor stays on
  step 1.
- What happens if the instructor has neither swim school nor coaching duties
  configured? The form shows a Swedish message indicating their profile is not
  set up for time reporting and prompts them to contact an administrator.
- What happens if the instructor presses the browser's Back button after
  reaching step 2? Step 1 is shown again with the previously entered email
  pre-filled for convenience.
- What happens with the name field (currently in the form)? The name field
  remains on step 2; the instructor still enters it manually since no name is
  stored in the backend.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The time report form MUST be divided into two sequential steps:
  email identification (step 1) and role-specific reporting (step 2).
- **FR-002**: Step 1 MUST present a single email input and a submit control.
  No other report fields are visible until the email is resolved.
- **FR-002b**: While the email lookup is in progress, the step 1 submit button
  MUST be disabled and show a loading indicator, preventing double-submission.
- **FR-003**: On step 1 submission, the system MUST look up the entered email
  against the instructor records in the backend.
- **FR-004**: If the email is not found or the backend is unreachable, step 2
  MUST NOT be shown; a clear Swedish error message MUST be displayed in step 1.
- **FR-005**: If the email is found, the system MUST advance to step 2 and
  display the sections determined by the instructor's configured roles.
- **FR-006**: Step 2 MUST show the Simskola section if and only if the
  instructor has swim school duties configured (`swim_school_rate` is set).
- **FR-007**: Step 2 MUST show the Träningsgrupper section if and only if the
  instructor has coaching duties configured (`coach_rate` is set).
- **FR-008**: Step 2 MUST show the Övrig tid (extra time) section for ALL
  instructors regardless of role combination.
- **FR-008b**: Step 2 MUST show the Kommentarer (comments) field for ALL
  instructors regardless of role combination.
- **FR-009**: Step 2 MUST show the Utlägg (expense) section if and only if the
  instructor has coaching duties configured (`coach_rate` is set). Swim
  school-only instructors MUST NOT see the Utlägg section.
- **FR-010**: The Milersättning (travel compensation) field MUST be shown in
  step 2 if and only if the instructor's profile has `travel_compensation`
  set to true.
- **FR-011**: The email address resolved in step 1 MUST be used as the
  submission email in step 2; the instructor MUST NOT need to re-enter it.
- **FR-012**: The Instructor entity in the backend MUST include a boolean
  attribute `travel_compensation` (default false) indicating eligibility for
  travel compensation.
- **FR-013**: The Instructor data model MUST support three role combinations:
  swim school only (`swim_school_rate` set, `coach_rate` NULL), coaching only
  (`coach_rate` set, `swim_school_rate` NULL), and both. At least one rate MUST
  be set for a valid instructor profile.
- **FR-014**: All user-visible messages, labels, and error texts MUST be in
  Swedish.
- **FR-015**: Step 2 MUST include the name field (instructor enters manually)
  since no display name is stored in the backend.
- **FR-016**: `pnpm build` MUST pass with zero TypeScript errors after all
  changes are applied.

### Key Entities

- **Instructor** (backend, personal data — admin-only): Identified by email
  address. Carries role flags (swim school duties, coaching duties) and a
  `travel_compensation` boolean. Also carries salary rates used elsewhere.
  The combination of role flags determines which form sections are shown in
  step 2.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An instructor can complete the email lookup in step 1 and reach
  the role-appropriate form in step 2 within 15 seconds on a standard mobile
  connection.
- **SC-002**: No training group sections are shown to a swim-school-only
  instructor, and no swim school section is shown to a coach-only instructor.
- **SC-003**: The Milersättning field is absent from the page DOM (not merely
  hidden via CSS) for instructors without `travel_compensation = true`.
- **SC-004**: An unknown email in step 1 produces a visible Swedish error
  message without advancing to step 2.
- **SC-005**: `pnpm build` completes with zero errors and zero `astro check`
  warnings after the feature is implemented.
- **SC-006**: The existing email delivery flow (submission → Mailjet → payroll
  inbox) continues to work for all role combinations — no regressions.

## Assumptions

- Email lookup in step 1 is a direct backend query (no one-time password or
  magic link is sent); the email is used as an identifier, not verified via
  delivery.
- The existing Instructor entity in Trailbase will be extended with a
  `travel_compensation` boolean column. Role determination follows the
  nullable rate columns: `swim_school_rate` set = swim school duties,
  `coach_rate` set = coaching duties.
- The `swim_school_rate` column must allow NULL to support coach-only
  instructors. The existing NOT NULL constraint on this column will be relaxed.
- The name field stays on step 2 because the instructor's display name is not
  stored in the backend (only their email and rates).
- The form continues to operate without requiring the instructor to log in
  (consistent with FR-009 from spec 002-time-report-trailbase).
- Övrig tid (extra time) is always visible in step 2.
- Utlägg (expense) is visible only when `coach_rate` is set (coaches and
  instructors with both roles).
- The step transition happens client-side (no full page reload); the backend
  lookup for step 1 is a dedicated API call made from the browser.

## Clarifications

### Session 2026-04-26

- Q: Which sections are visible per role combination? → A: Swim school only (`swim_school_rate` set, `coach_rate` NULL): Simskola + Övrig tid. Coaching only (`coach_rate` set, `swim_school_rate` NULL): Träningsgrupper + Övrig tid + Utlägg. Both roles: Simskola + Träningsgrupper + Övrig tid + Utlägg.
- Q: Is Övrig tid shown for all instructors? → A: Yes — all instructors see Övrig tid regardless of role.
- Q: Is Utlägg shown for swim-school-only instructors? → A: No — Utlägg is hidden when `coach_rate` is NULL.
- Q: Can `swim_school_rate` be NULL? → A: Yes — coach-only instructors have `swim_school_rate` NULL. The existing NOT NULL constraint will be relaxed.
- Q: Is the Kommentarer field visible for all instructors or role-dependent? → A: Always visible for all instructors (same as Övrig tid).
- Q: What feedback does the instructor see while the step 1 email lookup is in progress? → A: The submit button becomes disabled and shows a loading indicator; no other layout change.
