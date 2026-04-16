# Feature Specification: Time Report Trailbase Migration

**Feature Branch**: `002-time-report-trailbase`
**Created**: 2026-04-16
**Status**: Draft
**Input**: User description: "Migrate the current implementation of the time report functionality to utilize the new Trailbase backend. The time report does not require a login at this point. Take the opportunity to refactor and improve the code quality."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Admin Publishes Monthly Schedule (Priority: P1)

An administrator adds or updates the list of training sessions and swim school
occasions for the current reporting month in the admin interface. Instructors
visiting the time report form on alvestass.se immediately see the correct,
up-to-date schedule — no code change or deployment required.

**Why this priority**: Every month the schedule must be updated. Today this
requires a code deployment. Moving schedule data to the backend removes the
deployment dependency and unblocks non-technical administrators.

**Independent Test**: An administrator creates a new session entry in the
backend, then opens the time report form and confirms the new session appears
in the correct group list.

**Acceptance Scenarios**:

1. **Given** a new session exists in the backend for the active month,
   **When** an instructor opens the time report form,
   **Then** the new session appears in the correct group (simskola,
   tävlingsgrupp, etc.) with the correct date and title.

2. **Given** a session is removed from the backend,
   **When** an instructor opens the time report form,
   **Then** the removed session no longer appears in the list.

3. **Given** the backend is unreachable,
   **When** an instructor opens the time report form,
   **Then** a clear Swedish error message is shown and the form is not
   displayed in a broken or empty state.

---

### User Story 2 — Admin Updates Instructor Salary Rates (Priority: P2)

An administrator updates an instructor's hourly rate in the admin interface.
The next time report submitted by that instructor automatically uses the
updated rate in the preliminary salary estimate — without a code deployment
or a change to the repository.

**Why this priority**: Salary rates change with agreements, promotions, and
new hires. Today each change requires a code edit exposing personal data
(email addresses, salary rates) in the repository. Moving this data to the
backend removes personal data from source control and allows rate changes
without a deployment.

**Independent Test**: An administrator updates one instructor's swim school
rate in the backend, then submits a time report for that instructor's email
address and confirms the preliminary salary estimate reflects the new rate.

**Acceptance Scenarios**:

1. **Given** an instructor's swim school rate is updated in the backend,
   **When** that instructor's email address is entered in the time report form
   and a session is selected,
   **Then** the preliminary salary estimate in the email uses the updated
   hourly rate.

2. **Given** a new instructor is added to the backend,
   **When** that instructor submits a time report using their registered email,
   **Then** a salary estimate is included in the submitted report.

3. **Given** an email address is not registered in the backend,
   **When** that person submits a time report,
   **Then** the report is still submitted and delivered by email, but without
   a preliminary salary estimate (as today).

---

### User Story 3 — Admin Changes the Active Reporting Period (Priority: P3)

An administrator updates the active reporting month in the backend. The time
report form immediately reflects the new period — showing the correct month
name in the heading and loading the schedule for that month.

**Why this priority**: Changing the active month currently requires a code
change and deployment. This removes that friction but is lower priority than
the schedule and salary stories, which deliver more immediate operational
value.

**Independent Test**: An administrator sets a new active month in the backend
configuration, then opens the time report form and confirms the heading and
loaded sessions match the new month.

**Acceptance Scenarios**:

1. **Given** the active month is updated in the backend,
   **When** an instructor opens the time report form,
   **Then** the heading shows the correct new month name and the session
   list shows the sessions for that month.

2. **Given** no sessions exist in the backend for the configured active month,
   **When** an instructor opens the time report form,
   **Then** each group appears empty (no sessions to check off) and the
   form remains submittable.

---

### Edge Cases

- What happens when the backend returns an empty schedule for the active month?
  The form shows each group as empty and the instructor can still submit
  extra time and expense entries.
- What happens when the backend is temporarily unreachable when the form is
  loaded? A Swedish error message is shown; no broken or partially-loaded form
  is presented.
- What happens when an instructor submits a form while the backend is
  unreachable at submission time? The submission fails gracefully with a Swedish
  error message prompting the instructor to try again.
- What happens if the active month in configuration has no corresponding
  sessions in the schedule table? The form loads but groups are empty.
- What happens with expense (utlägg) attachments? File upload behaviour is
  unchanged; attachments are still sent by email.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The time report form MUST load its session schedule from the
  backend on each page render; no session data may be hardcoded in the
  deployed application code.
- **FR-002**: The time report form MUST display the active reporting month
  name as configured in the backend; the month setting MUST NOT be hardcoded
  in the deployed application code.
- **FR-003**: The backend MUST store instructor records containing at minimum:
  email address, swim school hourly rate, and coaching hourly rate (nullable).
- **FR-004**: Instructor records MUST be accessible only to authenticated
  backend administrators; they MUST NOT be readable via a public API endpoint.
- **FR-005**: The Cloudflare Worker handling form submission MUST look up
  the submitting instructor's rates from the backend using service credentials,
  NOT from hardcoded data in the source code.
- **FR-006**: The preliminary salary estimate included in the submitted
  email MUST use the rates retrieved from the backend at submission time.
- **FR-007**: Time report submissions MUST NOT store personal data (name,
  email, sessions, salary estimates) in the backend database; the email flow
  is the sole record of each submission.
- **FR-008**: Session schedule data MUST be readable by the Cloudflare Worker
  using the service user token (service-to-service call); no backend API
  endpoint is publicly accessible without a valid service token.
- **FR-009**: The time report form MUST remain publicly accessible without
  requiring the submitting instructor to log in.
- **FR-010**: If the backend is unreachable when the form is loaded, the page
  MUST display a clear Swedish error message rather than a broken form.
- **FR-011**: All configuration values that currently require a code change
  to update (active month, extra preparation times, special-day salary amounts)
  MUST be stored in the backend and retrieved at runtime.
- **FR-012**: The service credentials used for backend service-to-service calls
  MUST be stored as secrets in the Worker environment and MUST NOT appear in
  source code or be logged.

### Key Entities

- **Session** (schedule entry): Belongs to a reporting period (month key) and
  a group (simskola, tavlingA, tavlingB, teknik, masters, vuxencrawl). Has a
  date, title, duration in hours and minutes. Duration codes for full-day
  travel competitions (20 h), half-day (10 h) and overnight stay (15 h) must
  be preserved as-is.

- **Instructor** (personal data — admin-only): Email address (used as lookup
  key), swim school hourly rate (SEK/h), coaching hourly rate (SEK/h, nullable
  for instructors without coaching duties). Legal basis: employment/contractual
  necessity. Retention: until termination of employment + 1 year for accounting
  purposes.

- **TimeReportConfig** (single-row configuration): Active month key (e.g.
  "2026-04"), active month display name (Swedish, e.g. "april 2026"),
  extra preparation minutes for swim school sessions, extra preparation
  minutes for training sessions, half-day competition salary (SEK), full-day
  competition salary (SEK), overnight stay allowance (SEK).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An administrator can publish a new monthly schedule and have it
  visible on the time report form within 5 minutes, without any code change
  or deployment.
- **SC-002**: An administrator can update an instructor's hourly rate and have
  it reflected in the next submitted time report, without any code change or
  deployment.
- **SC-003**: No email addresses or salary rates appear in the application
  source code or repository after the migration.
- **SC-004**: The time report form loads and is fully usable within 3 seconds
  on a standard mobile connection.
- **SC-005**: `pnpm build` completes with zero TypeScript errors and zero
  `astro check` warnings after the migration.
- **SC-006**: The existing email delivery flow (submission → Mailjet → inbox)
  continues to work exactly as today — no regressions.

## Assumptions

- The form submission flow (Cloudflare Turnstile → Mailjet email) is
  unchanged; only the data sources migrate to Trailbase.
- Instructor records are managed exclusively through the Trailbase admin UI;
  no self-service instructor registration is in scope.
- A dedicated service user account will be created in Trailbase. The Worker
  authenticates as this service user (email + password, stored as Worker
  secrets) for ALL backend API calls — both schedule loading on page render
  and instructor rate lookup on form submission. No backend API endpoint is
  accessible without a valid service token.
- The schedule structure (six named groups: simskola, tavlingA, tavlingB,
  teknik, masters, vuxencrawl) remains fixed; adding new group types is out
  of scope.
- Expense (utlägg) file attachments continue to be sent by email; file
  storage in Trailbase is out of scope.
- The preliminary salary estimate logic (hours × rate + preparation time +
  competition allowances) is correct and is not in scope for business logic
  changes — only its data source changes.
- Backward compatibility with existing month data in `time-report-items.json`
  is not required; historical data will be entered into Trailbase as needed.

## Clarifications

### Session 2026-04-16

- Q: Should time report submissions be stored in the backend database in
  addition to being emailed? → A: No — email is the sole record. No personal
  submission data is stored in Trailbase.
- Q: What is the canonical term for the people who submit time reports? →
  A: "Instructor" — the term "employee" is retired. All references in this
  spec, the data model, and implementation use "instructor" exclusively.
- Q: What is the canonical term for the named groupings in the schedule
  (simskola, tävlingsgrupp, etc.)? → A: "group" — the term "section" is
  retired. All occurrences of "section" (referring to schedule groupings)
  have been replaced with "group" throughout this spec.
- Q: Should backend API endpoints for schedule data require authentication,
  or may they be publicly accessible? → A: All backend API endpoints MUST
  require the service user token; no endpoint is publicly accessible. The
  Worker uses the service credentials for every Trailbase call (schedule
  loading on page render and instructor rate lookup on submission).
