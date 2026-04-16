# Feature Specification: Trailbase Backend Setup (Minimal Starter)

**Feature Branch**: `001-trailbase-backend-setup`
**Created**: 2026-04-16
**Status**: Draft
**Input**: User description: "Setup the Trailbase backend as a minimal starter for the backend and future development. Store only a minimum amount of public club information there to begin with, which can be shown on the public facing website, as a proof of concept."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Public Visitor Sees Live Club Info (Priority: P1)

A person visiting alvestass.se sees up-to-date public club information — name,
tagline, founding year, short description, address, phone, and public email —
displayed on the website. This information is served from the backend, not
hard-coded in the source code.

**Why this priority**: This is the core proof-of-concept outcome. It proves that
the backend is running, connected, and serving real data to the public website.
Without this working, nothing else in this feature has value.

**Independent Test**: Open the website's contact or about page in a browser and
confirm that club name, address, phone, and email are displayed. Then update a
field in the backend admin interface and confirm the website reflects the change
within the expected refresh window — without any code change or redeployment.

**Acceptance Scenarios**:

1. **Given** the backend holds the club's contact details, **When** a visitor
   opens the relevant page, **Then** the page shows the club's current name,
   address, phone number, and public email.
2. **Given** an admin has just updated the club's tagline in the backend,
   **When** enough time has passed for the website to pick up the change,
   **Then** the new tagline is visible to all visitors without any code
   deployment.
3. **Given** the backend is temporarily unreachable, **When** a visitor opens
   the page, **Then** the page does not crash or display an unhandled error; it
   shows the last known club information rather than an error or blank section.

---

### User Story 2 - Admin Updates Club Info Without a Deployment (Priority: P2)

A developer or technical admin logs in to the backend admin interface and
edits any field of the public club information record. The change takes effect
on the website within a short, predictable window — no code change, no pull
request, no deployment needed. Non-technical board access is out of scope for
this PoC phase.

**Why this priority**: This demonstrates the core value proposition of having a
backend: content can be kept accurate without a code deployment. It validates
that the data-to-website pipeline works end-to-end.

**Independent Test**: Log in to the backend admin interface, change the club's
public phone number, wait for the refresh window to pass, and confirm the new
number appears on the website.

**Acceptance Scenarios**:

1. **Given** an admin is authenticated in the backend interface, **When** they
   edit and save the club's short description, **Then** the updated description
   appears on the website within 5 minutes and no code deployment is required.
2. **Given** an admin saves an empty value for a required field (e.g., club
   name), **Then** the system rejects the save and shows a validation error —
   the website continues to display the previous valid value.

---

### Edge Cases

- What happens when the backend is unreachable at page-render time? The page
  MUST display the last known cached club information rather than an error or
  blank section.
- What happens if a required field (club name, address) is cleared by an admin?
  The backend MUST reject the save with a validation error; the previous valid
  data MUST remain live on the website.
- What happens if the club information record does not exist yet (fresh
  deployment)? The website MUST handle a missing record gracefully — showing
  placeholder text rather than crashing.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The backend MUST store a single public club information record
  containing: club name, tagline, founding year, short description (≤ 300
  characters), postal address, city, postal code, public phone number, and
  public email address.
- **FR-002**: The club information record MUST be readable by the website
  without authentication — no login is required to fetch public club info.
- **FR-003**: An authorized administrator MUST be able to create, view, and
  update the club information record via the backend's admin interface without
  any code change or redeployment.
- **FR-004**: The website MUST display club information sourced from the
  backend; hard-coding these values in source files is not permitted once this
  feature is live.
- **FR-005**: All required fields (club name, address, email) MUST be validated
  as non-empty before a save is accepted. Saves with missing required fields
  MUST be rejected with a clear error.
- **FR-006**: The initial data set seeded into the backend MUST contain zero
  personal data — only publicly visible organizational information.
- **FR-007**: The backend MUST be deployed in a European data centre to satisfy
  GDPR data-residency requirements. The designated region is Stockholm (`arn`).

### Key Entities *(include if feature involves data)*

- **ClubInfo**: Represents the club's public identity and contact details.
  Single record (one club). Fields: name, tagline, founding_year,
  short_description, address, city, postal_code, phone, email.
  — **GDPR legal basis**: Legitimate interest — this is publicly available
  organizational contact information that the club is expected and required to
  share with the public. No personal data is involved in this initial scope.
  — **Retention**: Indefinite; this is organizational identity data, not
  personal data.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A visitor to the website can see the club's name, address, phone
  number, email, and short description, all sourced from the backend — verified
  by confirming the values match what is stored in the backend admin interface.
- **SC-002**: An admin can update any field of the club information and the
  change is visible on the website within 5 minutes, with no code deployment
  or developer intervention required.
- **SC-003**: The page displaying backend-sourced club information loads in
  under 3 seconds on a standard mobile connection.
- **SC-004**: Zero GDPR-classified personal data is present in the initial
  seeded data set (verified by inspecting all stored fields).
- **SC-005**: The backend withstands at least 100 consecutive read requests
  from the website without error or data corruption.

---

## Assumptions

- The club has a single publicly shareable contact email and phone number that
  the board has approved for publication on the website.
- The Trailbase admin interface is sufficient for all content management in
  this initial phase — no custom admin UI needs to be built as part of this
  feature.
- Only one club information record will ever exist; there is no need to manage
  multiple records or versioning in this phase.
- The Trailbase instance is hosted on the free tier of fly.io, deployed in the
  Stockholm region (`arn`) for GDPR data-residency compliance. Free-tier
  resource limits (shared CPU, 256 MB RAM, 3 GB persistent volume) are
  sufficient for this proof-of-concept load.
- No formal uptime SLA is required for this proof-of-concept; the free tier's
  best-effort availability is acceptable. The website's fallback behaviour
  (show last known cached values) mitigates brief outages.
- The admin interface is used exclusively by the developer / technical admin
  in this PoC phase. Non-technical board member access is explicitly out of
  scope and will be addressed in a future iteration. Authentication is handled
  entirely by Trailbase's built-in auth — no separate identity provider is
  needed for this feature.
- The website will display club info with a short delay (up to 5 minutes) after
  an admin update; real-time push is out of scope for this proof-of-concept.
- Data backup of the fly.io volume is out of scope for this PoC phase; the
  stored data is organizational contact info that can be re-entered if lost.

---

## Clarifications

### Session 2026-04-16

- Q: Should the fly.io deployment be required to run in a European region for
  GDPR data-residency compliance? → A: Yes — Stockholm region (`arn`).
- Q: When the backend is unreachable, what should the website show visitors?
  → A: Show the last successfully fetched club info (stale data, no visible
  error).
- Q: Who is expected to use the admin interface in this PoC phase?
  → A: Developer / technical admin only. Non-technical board member access is
  out of scope for this PoC.
