# Feature Specification: Member Register – Data Model and Initial Import

**Feature Branch**: `010-member-import`  
**Created**: 2026-05-19  
**Status**: Draft  
**Input**: User description: "The system needs to contain the currently active swimmers (and their guardians), instructors and board members so they can login and access relevant information, as well as update i.e attendance. A starting point establishing a data model and doing an initial import is needed."

## Clarifications

### Session 2026-05-19

- Q: Are "Ledare" and "Huvudledare" in WeUnite club members or employed instructors, and where should they be stored? → A: Instructors (Ledare/Huvudledare) are not necessarily club members — they may be purely employed. They appear in the WeUnite Grupplista because their attendance is tracked there. They should be stored in the existing `instructors` table, NOT in `members`, unless they are also active swimmers (e.g. a Masters swimmer who also leads a Simskola group). In that dual-role case they belong in both tables.
- Q: Should instructor-to-training-group assignments be recorded during this import? → A: No — instructor group links are deferred to the attendance/scheduling feature. Only swimmer memberships are recorded in `member_training_groups`. WeUnite "Ledare" rows are used only to cross-check existing `instructors` table entries and flag mismatches.
- Q: Which IdrottOnline roles qualify as "board member" for import into `members`? → A: Formal board positions only: Styrelseledamot, Ordförande, Vice ordförande, Kassör, Sekreterare. Administrative roles (Klubbadministratör, LOK-stödsansvarig, Kontakt dataskydd, Utbildningsansvarig) do not qualify.
- Q: What defines an "active" member for import purposes? → A: A person must appear in the WeUnite Grupplista with role "Deltagare" in any group — regardless of the group's end date — OR hold a formal board position in IdrottOnline. IdrottOnline membership dates (`Член t.o.m.`) are not used as an import filter; WeUnite Deltagare presence is the definitive criterion for swimmers.
- Q: Should the `members` table store the membership end date (`Член t.o.m.`)? → A: No — out of scope for this import.
- Q: Should role flags (is_swimmer, is_instructor) be stored as columns in the `members` table? → A: No. `is_swimmer` is unnecessary because the default for any member is that they are a swimmer — there is no need to store it explicitly. `is_instructor` is also not stored: instructors are employees managed in the separate `instructors` table; if one happens to also be a Deltagare they are an ordinary member with no special flag. Only `is_board_member` is stored, because board membership is a distinct sub-role that needs to be distinguishable. WeUnite "Ledare" rows are not cross-referenced against the `instructors` table and do not produce any import warnings.

## User Scenarios & Testing *(mandatory)*

### User Story 1 – Administrator imports active members (Priority: P1)

An administrator needs to perform an initial import of all persons who appear in the WeUnite Grupplista as "Deltagare" (in any group, regardless of end date), plus formal board members identified in IdrottOnline. Instructors who are not also Deltagare are reconciled against the existing `instructors` table and do not become `members` records. Historical members not appearing in WeUnite must not be imported.

**Why this priority**: Without a populated register, no other functionality (e.g. attendance tracking, login) can work. This is the foundation for the entire system.

**Independent Test**: Can be tested independently by running the import against the export files and verifying: (a) every Deltagare in WeUnite appears in `members`; (b) no person absent from WeUnite (and not a board member) appears in `members`; (c) instructors who are not Deltagare do not appear in `members`.

**Acceptance Scenarios**:

1. **Given** export files from IdrottOnline and WeUnite, **When** the import runs, **Then** records are created in `members` for every person with a "Deltagare" role in any WeUnite group, plus all formal board members from IdrottOnline — regardless of group end dates.
2. **Given** a person appears in the WeUnite export as "Deltagare" but has no matching IdrottOnline entry (no IID), **When** the import runs, **Then** that record is logged as a warning and skipped — an IID is required.
3. **Given** a person exists in both source systems, **When** the import runs, **Then** that person is created as a single record using the IID number as the identifier — no duplicate is created.
4. **Given** a member is under 18 years old, **When** the import runs, **Then** their guardians are also imported with contact details.
5. **Given** a training group is named "Baddaren 12.55–13.40" in the source data, **When** the import runs, **Then** the group is stored with only the name "Baddaren" — without the time slot.
6. **Given** an instructor also appears as "Deltagare" in a training group (e.g. a Masters swimmer who also leads Simskola), **When** the import runs, **Then** that person is added to `members` as a regular member — no special flag is set and no cross-reference warning is produced.

---

### User Story 2 – System identifies family connections (Priority: P2)

To support family discounts and reach the right guardians, the administrator needs to be able to see which members belong to the same family.

**Why this priority**: Family connections are important for future features (e.g. billing and communication) but are not a blocking requirement for the MVP import.

**Independent Test**: Can be tested independently by verifying that siblings and/or parents are correctly linked to a shared family unit in the database.

**Acceptance Scenarios**:

1. **Given** multiple active members in the source data share a family (per the IdrottOnline export), **When** the import runs, **Then** they are linked to a shared family unit in the database.
2. **Given** a guardian is registered for two siblings, **When** the system displays family information, **Then** the guardian is shown as connected to both children.

---

### User Story 3 – Administrator can validate import results (Priority: P3)

After a completed import, the administrator needs to be able to verify that the data is complete and correct — for example, by checking the count of imported records per category.

**Why this priority**: Important for data quality, but not a hard requirement to get started — the import can be re-run if errors are discovered.

**Independent Test**: Can be tested independently by displaying a summary of the import result (counts of members, board members, guardians, training groups) that the administrator can cross-check against the source files.

**Acceptance Scenarios**:

1. **Given** a completed import, **When** the administrator reviews the result, **Then** a summary is shown with the number of imported records per category (members, board members, guardians, training groups, family links).
2. **Given** a record is missing required data (e.g. no IID number), **When** the import runs, **Then** that record is logged as an error and skipped without aborting the entire import.

---

### Edge Cases

- What if a person is both a Deltagare and a board member? The person is stored as a single `members` record with `is_board_member = true`.
- What if a guardian is also a Deltagare in WeUnite? The guardian's IID number is used and they are linked as a guardian without being duplicated.
- What if a swim school group in the source data has no time slot in its name? The group is imported as-is.
- What if the same IID number appears in both source files with conflicting personal data? The IID is the master key — IdrottOnline is the primary source for personal data; WeUnite is the primary source for group membership.
- What if a Deltagare in WeUnite has no matching IdrottOnline entry (no IID)? The record is logged as a warning and skipped — an IID is required to create a `members` row.
- What if an instructor (Ledare/Huvudledare) in WeUnite has no matching entry in the existing `instructors` table? The instructor is logged as a warning — they are not added to `members`; the administrator should verify manually.
- What if an instructor also appears as "Deltagare" in a training group (dual swimmer-instructor role)? That person is imported into `members` as a regular member; no special flag is set. Their `instructors` table entry is unaffected and no cross-reference warning is produced.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST support a person register (`members` table) with the IID number (IdrottsID) as the unique identifier for each person.
- **FR-002**: The `members` table stores swimmers (Deltagare in WeUnite) and formal board members. The `is_board_member` flag distinguishes board members; no `is_swimmer` or `is_instructor` flags are stored — all members are swimmers by default, and instructor status is managed solely in the `instructors` table. A person who is both a Deltagare and an instructor is stored as a regular member with no special flag.
- **FR-003**: The system MUST store guardians linked to the active members they are responsible for.
- **FR-004**: A guardian MUST be able to have their own IID number if they are also registered in IdrottOnline, but IID is optional for guardians.
- **FR-005**: The system MUST store training groups without time slots in their names — times are handled in a separate scheduling step.
- **FR-006**: Training groups MUST be categorised as one of: swim school, adult, masters, competitive, technique.
- **FR-007**: The system MUST support a connection between a member and one or more training groups.
- **FR-008**: The system MUST support family connections so that members within the same family can be identified.
- **FR-009**: The import MUST include all persons who appear in the WeUnite Grupplista with role "Deltagare" in any group (regardless of the group's `Slut` date), plus all formal board members from IdrottOnline. Persons absent from both sources MUST be excluded.
- **FR-010**: The import MUST handle persons who exist in both source systems without creating duplicates, using the IID number as the deduplication key.
- **FR-011**: The import MUST log records that cannot be imported (e.g. Deltagare with no matching IID in IdrottOnline, instructor not in the `instructors` table) and continue processing the remaining records.
- **FR-012**: The system MUST store sufficient personal data to support future login functionality (name, email, IID).

### Key Entities

- **Member**: A person who appears in the WeUnite Grupplista as "Deltagare" in any group, or holds a formal board position. Uniquely identified by IID number. May hold both roles simultaneously. May belong to a family. Instructors who are not Deltagare are NOT members.
- **Instructor**: A person employed or volunteering to lead training groups. Stored in the existing `instructors` table (email + salary rates). Not in `members` unless they also appear as Deltagare in a WeUnite group.
- **Guardian**: A legal guardian linked to one or more active members. Has contact details (phone, email). May have their own IID number if registered in IdrottOnline.
- **Training Group**: A named group of swimmers with a category (swim school, adult, masters, competitive, technique). Time slots are stored separately in a future scheduling step.
- **Family**: A link that connects members in the same household or family unit, to support family discounts and grouped communication.
- **Group Membership**: The connection between a specific member and a training group, including their role in the group (participant, instructor, head instructor). Only `members` records can have a group membership entry.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every person who appears as "Deltagare" in any WeUnite group, and every formal board member from IdrottOnline, can be found and retrieved in the `members` table after a completed import.
- **SC-002**: No person absent from both WeUnite (as Deltagare) and the IdrottOnline board-member roles exists in the `members` table after the import.
- **SC-003**: No person appears as a duplicate — each IID number occurs exactly once in `members`.
- **SC-004**: All training group names in the database are free of time slots.
- **SC-005**: The import process can be re-run from scratch (idempotent) without creating duplicates.
- **SC-006**: The import process produces a report stating the number of imported records per category (members, board members, guardians, training groups, family links) and the number of skipped records with reasons.
- **SC-007**: Family connections exist for at least the family constellations visible in the IdrottOnline export.
- **SC-008**: Instructors (Ledare/Huvudledare) who do not appear as Deltagare in any WeUnite group do not exist in the `members` table after import.

## Assumptions

- "Active member" for this import is defined as: the person appears in the WeUnite Grupplista with role "Deltagare" in any group (regardless of the group's end date), OR holds a formal board position in IdrottOnline. IdrottOnline membership dates are not used as an import filter.
- "Formal board member" is defined as: the person holds one of the following roles in IdrottOnline without an end date: Styrelseledamot, Ordförande, Vice ordförande, Kassör, Sekreterare.
- Personal data (name, gender, date of birth, address, IID, email, phone) comes from IdrottOnline. Training group membership and guardian details come from WeUnite.
- Instructors (Ledare/Huvudledare in WeUnite) are managed in the existing `instructors` table. This import does not add new rows to `instructors` — it only cross-references existing entries.
- Swim school groups are categorised as "swim school"; A-gruppen/B-gruppen as "competitive"; Teknikgruppen as "technique"; Masters as "masters"; Vuxencrawl as "adult".
- Future login functionality is out of scope for this feature — the data model should be designed to support it.
- The import is performed manually by an administrator and does not need to run automatically or on a schedule at this stage.
- Personal data is handled in accordance with GDPR — no personal data in URLs or logs, and data is stored only in Trailbase behind authentication (planned in a future feature).
