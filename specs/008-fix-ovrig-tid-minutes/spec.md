# Feature Specification: Fix Övrig Tid Minutes Bug

**Feature Branch**: `008-fix-ovrig-tid-minutes`  
**Created**: 2026-04-30  
**Status**: Draft  
**Input**: User description: "Fix a bug where not typing anything into the minutes part of "Övrig tid" in the time report results in no reported time."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Report whole-hour extra time (Priority: P1)

An instructor fills in the "Övrig tid" section of the time report with a date, a number of hours, an empty minutes field (zero minutes), and a description, then submits the form. The row should appear in the submitted report and contribute to the salary calculation.

**Why this priority**: This is the most common case — reporting time in whole hours with no minutes. The current bug silently drops the entire row when minutes is left blank, meaning the instructor's time is lost and the salary estimate is wrong. This is the core defect.

**Independent Test**: Can be fully tested by submitting a time report with one Övrig tid row where the minutes field is left empty, and verifying the row appears in the confirmation email with the correct time.

**Acceptance Scenarios**:

1. **Given** an instructor has filled in Övrig tid with a valid date, 2 hours, empty minutes, and a description, **When** they submit the form, **Then** the row appears in the email with "2 h 0 m" and contributes 2 hours to the salary calculation.
2. **Given** an instructor fills in Övrig tid with 0 hours and 30 minutes, **When** they submit the form, **Then** the row appears in the email with "0 h 30 m".
3. **Given** an instructor fills in Övrig tid with 1 hour and 45 minutes, **When** they submit the form, **Then** the row appears in the email with "1 h 45 m".

---

### User Story 2 - Partial rows are handled gracefully (Priority: P2)

An instructor accidentally leaves an Övrig tid row completely empty (no date, no hours, no minutes, no description) and submits. That empty row should not appear in the submitted report.

**Why this priority**: The existing behaviour of silently ignoring incomplete rows was intentional for fully empty rows. The fix must not regress this behaviour.

**Independent Test**: Submit a form with one blank Övrig tid row and one properly filled row; verify only the filled row appears in the email.

**Acceptance Scenarios**:

1. **Given** an Övrig tid row where all fields are empty, **When** the form is submitted, **Then** that row does not appear in the email.
2. **Given** an Övrig tid row where only the description is missing, **When** the form is submitted, **Then** the row is treated as incomplete and is not included in the report.

---

### Edge Cases

- What happens when the hours field is also left empty but minutes has a value? (e.g., 0 h, 45 m)
- What happens when an instructor enters 0 for both hours and minutes but provides a date and description? (row should be included with 0 h 0 m, as the instructor explicitly reported it)
- What happens if hours or minutes contain non-numeric input?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A valid Övrig tid row MUST be included in the time report even when the minutes field is left empty (treated as 0 minutes).
- **FR-002**: A valid Övrig tid row MUST be included in the time report even when the hours field is left empty (treated as 0 hours).
- **FR-003**: A row is considered valid when it has at least a date, a description, and either a non-empty hours or minutes value (or both explicitly set to 0 but with a date and description present).
- **FR-004**: A row where all fields are empty MUST NOT be included in the submitted report.
- **FR-005**: A row missing a required identifying field (date or description) MUST NOT be included in the submitted report.
- **FR-006**: The salary calculation for Övrig tid MUST correctly account for hours and minutes from all valid rows, including those where minutes was left blank.
- **FR-007**: The email representation of an Övrig tid row MUST display "0" for the minutes when the field was left empty.

### Key Entities

- **Övrig tid row**: A user-entered time entry with fields: date, hours (h), minutes (m), and description. Hours and minutes default to 0 when blank.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An Övrig tid row submitted with an empty minutes field appears in the resulting email with the correct hours and 0 minutes — 100% of the time.
- **SC-002**: Salary calculation for Övrig tid matches the manually expected value when minutes is omitted — 0 discrepancy.
- **SC-003**: Completely empty Övrig tid rows never appear in the submitted report — 0 false positives.
- **SC-004**: All three acceptance scenarios in User Story 1 pass without manual workaround.

## Assumptions

- The fix is contained entirely in the form-data parsing logic (`timeReportValidation.ts`). No changes to the UI form fields, email templates, or salary calculation logic are expected.
- "Left empty" means the browser submits an empty string `""` for that field, not a missing key.
- A row with a date and description but 0 h 0 m is a valid submission — the instructor may be documenting presence without compensated time.
- Non-numeric values in hours or minutes fields are out of scope for this bug fix; existing behaviour (which ignores such rows) is acceptable.
