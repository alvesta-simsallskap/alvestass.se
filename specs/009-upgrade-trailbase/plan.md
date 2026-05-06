# Implementation Plan: Upgrade Trailbase to v0.26.8

**Branch**: `009-upgrade-trailbase` | **Date**: 2026-05-05 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/009-upgrade-trailbase/spec.md`

## Summary

Upgrade the Trailbase backend service from v0.26.3 to v0.26.8 and update the Go client SDK in the admin CLI to the corresponding latest snapshot. The Dockerfile already uses `trailbase/trailbase:latest`, so the server upgrade requires only a `fly deploy` redeploy. The Go client SDK in `tools/admin-cli/go.mod` is pinned to a specific commit and must be updated to `v0.0.0-20260501081523-08228bb11ccf`. Version references in `CLAUDE.md` must be updated. No schema migrations, no API contract changes, and no frontend changes are required.

## Technical Context

**Language/Version**: Go 1.24.4 (admin CLI); TypeScript 5.9 / Astro 5.17.1 (frontend — unaffected)  
**Primary Dependencies**: `github.com/trailbaseio/trailbase/client/go/trailbase` (Go SDK); `trailbase/trailbase:latest` Docker image  
**Storage**: Trailbase SQLite on fly.io (arn) — volume `trailbase_data` persists across redeploys  
**Testing**: Manual end-to-end test of time-report submission; `go build ./...` for admin CLI  
**Target Platform**: fly.io (Trailbase server); darwin/linux (admin CLI binary)  
**Project Type**: Maintenance / dependency upgrade — no new source files  
**Performance Goals**: Unchanged from current; no regressions expected  
**Constraints**: Zero downtime preferred; fly.io rolling deploy achieves this by default  
**Scale/Scope**: Single-instance Trailbase on fly.io; one admin CLI module

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ PASS | No new code introduced; existing type safety unaffected |
| II. Testing Standards | ✅ PASS | `pnpm build` unaffected (no frontend changes); `go build` must pass after SDK update |
| III. UX Consistency | ✅ PASS | No UI changes |
| IV. Performance | ✅ PASS | No frontend changes; server performance neutral |
| V. Backend Architecture | ✅ PASS | Trailbase remains the sole backend; this is a version bump |
| VI. External Data Integration | ✅ PASS | Not applicable |
| VII. GDPR | ✅ PASS | No new personal data; no new tables; no new endpoints |

**GATE RESULT: All gates pass. No violations.**

## Project Structure

### Documentation (this feature)

```text
specs/009-upgrade-trailbase/
├── plan.md              # This file
├── research.md          # Phase 0 output
└── tasks.md             # Phase 2 output (/speckit-tasks — not created here)
```

### Source Code (files touched by this feature)

```text
trailbase/
└── Dockerfile           # Already uses :latest — no change needed

tools/admin-cli/
├── go.mod               # SDK version bump: → v0.0.0-20260501081523-08228bb11ccf
└── go.sum               # Updated by go get

CLAUDE.md                # Version references: 0.26.3 → 0.26.8
```

**Structure Decision**: No new directories or files. This is a minimal maintenance change touching three files.

---

## Phase 0: Research

*See [research.md](research.md) for full findings.*

### Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Server upgrade mechanism | `fly deploy` (no Dockerfile change) | Dockerfile already uses `:latest`; redeploy pulls the new image |
| Go SDK version to use | `v0.0.0-20260501081523-08228bb11ccf` | Latest `main` commit as of 2026-05-01; corresponds to post-v0.26.8 SDK |
| Migration required? | No | No schema changes in v0.26.4–v0.26.8 |
| Breaking changes? | None | All releases v0.26.4–v0.26.8 are backwards-compatible |
| Data backup before upgrade? | Fly.io volume snapshot (optional best practice) | Volume persists independently; no data at risk from a Docker image swap |

---

## Phase 1: Design

### Data Model

No data model changes. No new Trailbase tables or schema migrations are required. The upgrade is a pure binary replacement.

### Contracts

No API contract changes. The Trailbase REST API used by `src/lib/trailbase.ts` is unchanged between v0.26.3 and v0.26.8.

### Implementation Steps (for task generation)

1. **Verify Dockerfile**: Confirm `trailbase/Dockerfile` uses `trailbase/trailbase:latest` (already confirmed — no change needed).
2. **Update Go client SDK**: `go.mod` already updated to `v0.0.0-20260501081523-08228bb11ccf` via `go get`; verify `go.sum` is also updated.
3. **Build admin CLI**: Run `go build ./...` inside `tools/admin-cli/` to confirm it compiles cleanly.
4. **Run admin CLI tests**: Run `go test ./...` inside `tools/admin-cli/` to confirm no regressions.
5. **Deploy Trailbase**: Run `fly deploy` from `trailbase/` to redeploy the service on fly.io.
6. **Verify server version**: Confirm v0.26.8 is running via the Trailbase admin UI.
7. **End-to-end smoke test**: Submit a time report to verify the full workflow against the upgraded server.
8. **Update CLAUDE.md**: Replace all references to `0.26.3` with `0.26.8`.

### Quickstart

```bash
# 1. Admin CLI — verify build and tests
cd tools/admin-cli
go build ./...
go test ./...

# 2. Deploy updated Trailbase service
cd ../../trailbase
fly deploy

# 3. Verify running version in Trailbase admin UI
# https://alvestass-trailbase.fly.dev/_/admin

# 4. Update CLAUDE.md version references
# sed -i '' 's/0\.26\.3/0.26.8/g' ../CLAUDE.md   (or edit manually)
```
