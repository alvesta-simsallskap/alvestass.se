# Tasks: Upgrade Trailbase to v0.26.8

**Input**: Design documents from `specs/009-upgrade-trailbase/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅

**Note**: No tests were requested in the spec. The Go client SDK (`go.mod`) was already updated to `v0.0.0-20260501081523-08228bb11ccf` during planning research; tasks below verify and complete that work.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1 = server upgrade, US2 = Go SDK / admin CLI)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm prerequisites and repository state before any changes are made

- [x] T001 Verify `trailbase/Dockerfile` still reads `FROM trailbase/trailbase:latest` (no edit needed — just confirm no accidental pin)
- [x] T002 Confirm `tools/admin-cli/go.mod` shows `github.com/trailbaseio/trailbase/client/go/trailbase v0.0.0-20260501081523-08228bb11ccf` (updated during planning; verify)
- [x] T003 Confirm `tools/admin-cli/go.sum` contains an entry for the new SDK commit (updated automatically by `go get` during planning; verify)

**Checkpoint**: Preconditions confirmed — proceed to story phases

---

## Phase 2: User Story 2 — Admin CLI Uses Updated Go SDK (Priority: P2)

**Goal**: Verify the admin CLI compiles and all existing tests pass against the updated Go client SDK.

**Independent Test**: Run `go build ./...` and `go test ./...` inside `tools/admin-cli/` — both must exit 0.

### Implementation

- [x] T004 [US2] Build the admin CLI to confirm the updated SDK compiles without errors: run `go build ./...` in `tools/admin-cli/`
- [x] T005 [US2] Run the admin CLI test suite to confirm no regressions: run `go test ./...` in `tools/admin-cli/`

**Checkpoint**: Admin CLI builds and tests pass — US2 complete

---

## Phase 3: User Story 1 — Trailbase Server Running Latest Version (Priority: P1)

**Goal**: Deploy the updated Trailbase image to fly.io and confirm the live service runs v0.26.8.

**Independent Test**: Visit the Trailbase admin UI at `https://alvestass-trailbase.fly.dev/_/admin` and confirm the reported version is v0.26.8. Then submit a time report to verify end-to-end functionality.

### Implementation

- [x] T006 [US1] Deploy the updated Trailbase service: run `fly deploy` from `trailbase/`
- [x] T007 [US1] Verify the running server version in the Trailbase admin UI (`https://alvestass-trailbase.fly.dev/_/admin`) — confirm v0.26.8 is reported
- [x] T008 [US1] End-to-end smoke test: submit a time report through the website and confirm it is stored correctly in Trailbase

**Checkpoint**: Live Trailbase service is on v0.26.8 and end-to-end workflow verified — US1 complete

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates and final build verification

- [x] T009 Update all `0.26.3` version references in `CLAUDE.md` to `0.26.8`
- [x] T010 [P] Run `pnpm build` from the repo root to confirm the Astro/Cloudflare build is unaffected
- [x] T011 Mark all tasks complete in tasks.md and commit the spec artifacts

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (US2)**: Depends on Phase 1 (go.mod verified) — can start immediately after T002/T003
- **Phase 3 (US1)**: Depends on Phase 1 only — independent of Phase 2; fly deploy does not require Go SDK changes
- **Phase 4 (Polish)**: Depends on US1 + US2 both complete

### User Story Dependencies

- **US2 (P2)**: Depends only on Phase 1; can run before or in parallel with US1
- **US1 (P1)**: Depends only on Phase 1; the fly.io deploy is independent of the Go SDK changes

### Parallel Opportunities

- T004 and T006 can run in parallel (Go build vs. fly deploy — completely different targets)
- T009 and T010 can run in parallel (different files)

---

## Parallel Execution Example

```bash
# After Phase 1 completes (T001–T003), launch in parallel:

# Terminal A — US2 (admin CLI)
cd tools/admin-cli
go build ./...     # T004
go test ./...      # T005

# Terminal B — US1 (Trailbase server)
cd trailbase
fly deploy         # T006 — takes ~2 minutes
# then verify via admin UI (T007) and smoke test (T008)
```

---

## Implementation Strategy

### MVP (User Story 1 — server upgrade)

1. Complete Phase 1: verify Dockerfile + go.mod state (T001–T003)
2. Complete Phase 3: deploy and verify (T006–T008)
3. **STOP and VALIDATE**: live service on v0.26.8, smoke test passes
4. Then add Phase 2 (admin CLI build verification) and Phase 4 (docs)

### Full delivery

1. Phase 1 → confirm preconditions
2. Phase 2 + Phase 3 in parallel → US2 and US1 simultaneously
3. Phase 4 → Polish and commit

---

## Notes

- [P] tasks = different files, no dependencies between them
- The `go.mod` and `go.sum` changes were made during the `/speckit-plan` research step — T002/T003 are verification tasks, not edit tasks
- No schema migrations are required
- No frontend (`src/`) changes are required
- `pnpm build` is the merge gate (Principle II of constitution) — T010 confirms it still passes
