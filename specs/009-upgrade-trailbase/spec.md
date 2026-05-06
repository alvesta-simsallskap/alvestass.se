# Feature Specification: Upgrade Trailbase to Latest Version

**Feature Branch**: `009-upgrade-trailbase`  
**Created**: 2026-05-05  
**Status**: Draft  
**Input**: User description: "Upgrade Trailbase so that the system is kept up to date."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Trailbase Server Running Latest Version (Priority: P1)

The deployed Trailbase service on fly.io runs the latest stable version (v0.26.8) instead of v0.26.3. Operators can confirm the running version through the Trailbase admin UI or health endpoint.

**Why this priority**: Running outdated server software accumulates security debt and delays access to bug fixes. This is the primary deliverable of the upgrade.

**Independent Test**: Can be fully tested by deploying the updated Docker image to fly.io and visiting the Trailbase admin UI to confirm the version number shown is v0.26.8.

**Acceptance Scenarios**:

1. **Given** the Trailbase service is deployed on fly.io, **When** an operator visits the Trailbase admin UI, **Then** the reported server version is v0.26.8 or later.
2. **Given** the upgraded service is running, **When** the time-report submission flow is exercised end-to-end, **Then** all existing functionality continues to work without errors.
3. **Given** the upgraded service is running, **When** the admin CLI performs read/write operations against Trailbase, **Then** all operations complete successfully.

---

### User Story 2 - Admin CLI Uses Updated Go Client SDK (Priority: P2)

The admin CLI Go module references a Go client SDK version that is compatible with and tested against Trailbase v0.26.8, replacing the current snapshot pinned to an older commit.

**Why this priority**: The Go client SDK is a direct dependency of the admin CLI. Keeping it in sync with the server version prevents subtle API mismatches and ensures the CLI benefits from any SDK fixes.

**Independent Test**: Can be tested by running the admin CLI against the upgraded Trailbase server and confirming all commands (list sessions, import CSV) work correctly.

**Acceptance Scenarios**:

1. **Given** the admin CLI go.mod is updated, **When** `go build ./...` is run inside `tools/admin-cli/`, **Then** the build completes without errors.
2. **Given** the updated CLI binary, **When** the CLI connects to the upgraded Trailbase server, **Then** authentication and all record operations succeed.

---

### Edge Cases

- What happens if the fly.io deployment fails mid-upgrade? The existing image remains active; the volume-mounted data is untouched.
- What if the new Trailbase server version changes the SQLite schema format? No schema changes occurred between v0.26.3 and v0.26.8 — existing migrations remain valid.
- What if the Go SDK module path or import path changed? Release notes confirmed no Go client changes — module path is unchanged.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Trailbase Docker image MUST be updated to reference v0.26.8 (or `latest` if already using it, verified to resolve to v0.26.8).
- **FR-002**: The fly.io deployment MUST be redeployed so the running service uses the new image.
- **FR-003**: The Go client SDK dependency in `tools/admin-cli/go.mod` MUST be updated to the latest release compatible with Trailbase v0.26.8.
- **FR-004**: The build (`go build ./...`) for the admin CLI MUST pass after the dependency update.
- **FR-005**: The `pnpm build` gate for the Astro site MUST continue to pass after the upgrade (no frontend Trailbase client changes required, but regression must be confirmed).
- **FR-006**: Version references in `CLAUDE.md` MUST be updated to reflect the new Trailbase version.

### Key Entities

- **Trailbase server**: The backend service running on fly.io; currently v0.26.3, target v0.26.8.
- **Docker image**: `trailbase/trailbase:latest` in `trailbase/Dockerfile` — currently uses `latest` tag (already points to newest, but the running instance needs redeployment).
- **Go client SDK**: `github.com/trailbaseio/trailbase/client/go/trailbase` in `tools/admin-cli/go.mod` — pinned to a commit from 2026-04-21; must be updated to latest.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The live Trailbase service reports version v0.26.8 in its admin UI within one deployment cycle.
- **SC-002**: The admin CLI build completes successfully with zero errors after the Go SDK update.
- **SC-003**: All existing time-report submission and admin CLI workflows complete without regression after the upgrade.
- **SC-004**: `CLAUDE.md` version references are updated to v0.26.8 within the same commit as the dependency changes.

## Assumptions

- The `trailbase/trailbase:latest` Docker tag already resolves to v0.26.8; redeploying the fly.io app is sufficient to pick up the new server version.
- There are no breaking changes between v0.26.3 and v0.26.8 (confirmed by release note review — all intermediate releases are backwards-compatible).
- The Trailbase Go client SDK is distributed as part of the monorepo; the latest module commit corresponds to v0.26.8.
- No SQLite migration is required — the upgrade is a pure binary replacement with no schema changes.
- The fly.io volume (`trailbase_data`) is preserved across redeployments, so no data migration or backup procedure is required beyond standard operational caution.
- Mobile/web frontend code does not import the Go client SDK; only `tools/admin-cli/` requires the Go dependency update.
