# Feature Specification: Admin CLI

**Feature ID**: 003  
**Short Name**: admin-cli  
**Status**: Draft  
**Created**: 2026-04-22  
**Author**: Johan Marand

---

## Overview

A standalone command-line interface (CLI) tool for administrators to manage backend data for the Alvesta Simsällskap website. The tool runs locally on a workstation and provides a simple, menu-driven interface to update and validate backend data without requiring direct database access or technical knowledge beyond running an executable.

---

## Problem Statement

Administrators currently have no self-service way to maintain backend data (such as club contact details). All data updates require either direct database access or developer involvement. This creates a bottleneck, slows down operations, and introduces risk of human error.

---

## Goals

- Allow administrators to perform common backend data operations without developer assistance
- Reduce time to update or correct backend data from hours/days to minutes
- Provide clear feedback and validation so administrators know immediately if something went wrong
- Work on administrator workstations regardless of operating system

---

## Non-Goals

- Does not replace or bypass security controls on the backend
- Does not expose raw SQL or direct database access
- Does not provide a web or graphical interface — CLI only
- Does not handle real-time monitoring or alerting
- Does not support importing data from files — the club info record always exists and is edited interactively

---

## User Scenarios & Testing

### Scenario 1: Update existing record
**Given** an administrator needs to correct a contact email or update the club's address  
**When** they select "Uppdatera kontaktuppgifter" and follow the prompts  
**Then** they can review current values, edit one or more fields, and confirm the update  
**And** the CLI confirms the update was applied successfully

### Scenario 2: Check for inconsistencies
**Given** an administrator suspects data quality issues  
**When** they select "Kontrollera data"  
**Then** the CLI runs validation checks and lists any problems found (missing required fields, invalid formats)  
**And** each problem includes enough context (field name, current value, rule violated) to know what to fix

### Scenario 3: View help
**Given** an administrator is unsure how to use a feature  
**When** they select "Hjälp" from the main menu  
**Then** they see a description of available operations and step-by-step instructions for each

### Scenario 4: Recover from error
**Given** an administrator enters an invalid value during an update  
**When** the CLI detects the error  
**Then** it displays a clear, plain-language error message in Swedish and allows the administrator to retry or cancel

---

## Functional Requirements

### Core Operations

1. **FR-01**: The CLI must present a main menu on launch listing all available operations, with numbered options selectable by keyboard.
2. **FR-02**: The CLI must provide an update operation that displays current field values and allows the administrator to modify one or more fields interactively.
3. **FR-03**: The CLI must provide a consistency check operation that fetches the current record, runs validation rules, and reports each violation with field name, current value, and rule description in Swedish.
4. **FR-04**: The CLI must provide a help screen accessible from the main menu, describing what each operation does and how to use it.
5. **FR-05**: The CLI must display a progress indicator during network operations (fetch, save).
6. **FR-06**: After each operation completes, the CLI must display a confirmation message or a summary of any errors.
7. **FR-07**: The CLI must support cancellation of any in-progress operation without corrupting backend data.

### Platform & Distribution

8. **FR-08**: The CLI must be distributed as a single self-contained executable file with no installation required.
9. **FR-09**: The CLI must run on macOS (Apple Silicon and Intel) and Windows 10/11 (64-bit) without installing any runtime or dependencies.

### Configuration & Authentication

10. **FR-10**: On first launch, the CLI must prompt the administrator to enter the backend URL, email, and password. The resulting session token must be stored locally for subsequent runs.
11. **FR-11**: Stored credentials must not be written to any shared or version-controlled location; they must be stored in a per-user configuration directory.
12. **FR-12**: The CLI must verify connectivity to the backend at startup and display a clear Swedish message if the backend is unreachable.

### Data Safety

13. **FR-13**: The update operation must validate all field values before sending any changes to the backend; validation errors must be shown before the confirm prompt.

---

## Success Criteria

1. An administrator with no prior training can launch the CLI, navigate to an operation, and complete it successfully within 10 minutes on first use.
2. The consistency check operation identifies 100% of missing required fields and format violations present in the current record.
3. Zero data corruption incidents occur when an operation is cancelled mid-way.
4. The CLI runs without error on macOS 14+ and Windows 10/11 without any additional software installation.
5. All user-visible text is written in plain Swedish and includes enough context for the administrator to understand what went wrong and what to do next.

---

## Key Entities

| Entity | Description |
|--------|-------------|
| Club info record | The single organizational contact record stored in Trailbase (`club_info`, id=1) |
| Configuration | Locally stored backend URL and session token used by the CLI to authenticate |
| Operation | A discrete task the CLI can perform (update, check, help) |
| Check issue / `CheckIssue` | A specific problem found in a field during update or consistency check |

---

## Assumptions

- The `club_info` record always exists (seeded by the initial migration); there is no create or delete operation.
- Administrators are comfortable running a terminal/command prompt at a basic level.
- The backend exposes REST endpoints to read and update the `club_info` record.
- Authentication uses a session token obtained by email/password login; the token is cached locally and refreshed on expiry.
- Swedish is the language for all user-visible text, consistent with the rest of the site.

---

## Dependencies

- Trailbase backend (fly.io, Stockholm region) must expose REST endpoints to read and update `club_info`
- Administrator workstation must have internet access to reach the backend

---

## Open Questions / Risks

- **Scope (resolved)**: The first version covers **contact information only**. Other entities are out of scope and may be added in a future iteration.
- **Import removed (resolved)**: CSV import is not needed — the `club_info` record is always seeded by the migration and only needs interactive updates.
