# Feature Specification: CSV Import of Time Report Sessions

**Feature Branch**: `006-import-sessions-csv`
**Created**: 2026-04-26
**Status**: Draft
**Input**: User description: "The CLI needs functionality to import time report sessions, preferrably as csv. In the import the combination of month_key, training_group, date and title should serve as a key so that matching existing rows gets updated, and non existing rows get inserted."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Import a New Session Schedule for a Month (Priority: P1)

An administrator has prepared a CSV file listing all training sessions for an
upcoming month. They run the import command from the CLI and point it at the
CSV file. The CLI reads the file, shows a summary of how many rows will be
inserted, and asks for confirmation. After the administrator confirms, the CLI
uploads all rows to the backend and reports success.

**Why this priority**: This is the primary motivation for the feature — bulk
population of sessions for a new month is far faster via CSV import than
manual entry, and it is the most common operation the administrator will
perform.

**Independent Test**: Create a CSV with three session rows (different
training_group values, all new). Run the import command. Confirm the CLI
shows "3 inserts, 0 updates" before confirmation. Confirm, then verify via the
CLI or backend that all three rows exist in the database.

**Acceptance Scenarios**:

1. **Given** a valid CSV file where none of the rows match existing records,
   **When** the administrator runs the import command with that file,
   **Then** the CLI displays a summary showing the number of rows to be
   inserted (0 updates), asks for confirmation, and upon approval inserts all
   rows and reports success.

2. **Given** the administrator has not yet confirmed the import,
   **When** they choose to cancel at the confirmation prompt,
   **Then** no changes are made to the backend and the CLI exits cleanly.

---

### User Story 2 — Update an Existing Session (Priority: P2)

An administrator needs to correct the duration of a session that was already
imported. They update the hours and minutes in the CSV file and re-run the
import. The CLI detects that the row (identified by month_key + training_group
+ date + title) already exists and shows it as an update, not an insert. After
confirmation, the existing record in the backend is overwritten with the new
values.

**Why this priority**: Without update support, correcting a session requires
manual deletion in the backend admin UI before re-importing. The upsert
behaviour makes the import idempotent and safe to re-run.

**Independent Test**: Import a CSV with one session row. Then change the hours
value in the CSV and re-run. Confirm the CLI shows "0 inserts, 1 update".
Confirm, and verify the backend record reflects the new hours.

**Acceptance Scenarios**:

1. **Given** a session with the same (month_key, training_group, date, title)
   already exists in the backend,
   **When** the administrator imports a CSV row with the same key but different
   hours or minutes,
   **Then** the CLI classifies the row as an update and shows "1 update" in
   the summary.

2. **Given** the administrator confirms the import containing both inserts and
   updates,
   **When** the CLI applies changes,
   **Then** new rows are inserted, existing rows are overwritten with the new
   hours and minutes, and unchanged rows (same key, same hours/minutes) are
   reported as skipped.

---

### User Story 3 — Receive Helpful Errors for an Invalid File (Priority: P3)

An administrator accidentally provides a malformed CSV or a file with an
unrecognised training_group value. The CLI validates the file before showing
the confirmation summary, lists every row-level error with its line number and
a plain-language description, and exits without making any changes.

**Why this priority**: Validation errors are caught before any data reaches the
backend, preventing partial imports and confusing database states.

**Independent Test**: Create a CSV with a row where training_group is
"ungdomA" (not a valid value) and another where hours is negative. Run the
import. Confirm the CLI lists both errors with their line numbers and makes no
backend changes.

**Acceptance Scenarios**:

1. **Given** a CSV file where one row has an invalid training_group value,
   **When** the administrator runs the import,
   **Then** the CLI lists the error with the line number and the invalid value,
   and exits without importing anything.

2. **Given** a CSV file that is missing the required header row or a required
   column,
   **When** the administrator runs the import,
   **Then** the CLI displays a clear error describing which column is missing
   and exits without importing anything.

3. **Given** a CSV file where the hours field is negative,
   **When** the administrator runs the import,
   **Then** the CLI reports a validation error for that row and exits without
   importing anything.

---

### Edge Cases

- What happens when the CSV file is empty (header only, no data rows)? → CLI
  reports "0 rows found" and exits without prompting for confirmation.
- What happens when the CSV contains duplicate composite keys within the same
  file? → CLI treats each occurrence as an upsert; last occurrence wins.
- What happens if the backend is unreachable during import? → CLI reports a
  connectivity error in Swedish showing how many rows were successfully applied
  before the failure, then exits cleanly. Since rows are applied one at a time
  with no transaction, a mid-import failure may leave a partial import; the
  administrator can correct this by fixing the issue and re-running the import
  (upsert semantics ensure already-applied rows are simply skipped on re-run).
- What happens when minutes is omitted from the CSV? → Defaults to 0.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide an import option in the main menu that
  prompts the administrator to enter the path to a CSV file via an interactive
  text prompt.
- **FR-002**: The CLI MUST parse the CSV file and validate every row before
  sending any changes to the backend.
- **FR-003**: The CLI MUST use the combination of (month_key, training_group,
  date, title) as a composite key to determine whether each row is an insert or
  an update.
- **FR-004**: The CLI MUST display a pre-confirmation summary showing the
  count of rows to insert, rows to update, and rows to skip (unchanged key +
  values), and then wait for explicit user confirmation before applying changes.
- **FR-005**: The CLI MUST validate that training_group is one of the
  permitted values: simskola, tavlingA, tavlingB, teknik, masters, vuxencrawl.
- **FR-006**: The CLI MUST validate that hours is a non-negative integer.
- **FR-007**: The CLI MUST validate that minutes is a non-negative integer less
  than 60; if the column is absent from the file it MUST default to 0.
- **FR-008**: The CLI MUST validate that month_key, date, and title are
  non-empty strings.
- **FR-009**: If any row fails validation, the CLI MUST list all errors with
  their line number and a plain-language Swedish message, and MUST NOT import
  anything.
- **FR-010**: After user confirmation, the CLI MUST apply inserts and updates
  and report the final count of each applied operation.
- **FR-011**: The CLI MUST display a progress indicator while fetching existing
  records and while uploading changes.
- **FR-012**: The CLI MUST be accessible from the existing main menu as a new
  option alongside the current operations.
- **FR-013**: The CSV file MUST include a header row; the CLI MUST accept
  columns in any order as long as all required columns are present.

### Key Entities

- **TimeReportSession**: A single training or swim-school session defined by
  (month_key, training_group, date, title, hours, minutes). The composite key
  (month_key + training_group + date + title) uniquely identifies a session
  within the import.
- **ImportResult**: The outcome of an import run, containing counts of
  inserted, updated, and skipped rows.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An administrator can import a correctly formatted CSV file of
  20 sessions in under 30 seconds from running the command to receiving the
  success confirmation.
- **SC-002**: 100% of validation errors in a CSV file are caught and reported
  before any row is sent to the backend.
- **SC-003**: Re-running an import with an identical CSV produces zero inserts
  and zero updates (all rows skipped as unchanged), confirming idempotency.
- **SC-004**: An administrator with no prior use of the import feature can
  successfully complete their first import without consulting external
  documentation, relying only on the CLI's built-in prompts and help text.

## Assumptions

- The CSV delimiter is a comma (`,`); quoted fields may contain commas.
- The `date` field in the CSV uses the format `YYYY-MM-DD`; the CLI does not
  enforce this format beyond requiring the field to be non-empty (consistent
  with how the database stores it as TEXT).
- The backend Trailbase API supports individual record creation and update
  operations; the CLI calls these per-row (no bulk endpoint assumed).
- The administrator runs the import from a workstation with access to the local
  CSV file and connectivity to the Trailbase backend.
- The import feature targets the same `time_report_sessions` table that the
  time report wizard reads from; no schema changes are required.
- Authentication uses the existing session token stored by the CLI; no new
  authentication mechanism is needed.
