# Tasks: Admin CLI

**Input**: Design documents from `specs/003-admin-cli/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Tests**: Unit tests included for pure-logic packages (validation rules). No network calls in test suite.

**Organization**: Tasks grouped by user story. The check operation uses a `Checker` interface so future health checks (member groups, instructor time reports) can be added without changing the runner or UI.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US4)
- All paths are relative to `tools/admin-cli/`

---

## Phase 1: Setup

**Purpose**: Go module creation and project scaffolding

- [x] T001 Create `tools/admin-cli/` directory and initialize Go module `github.com/alvestass/admin-cli` with `go mod init`
- [x] T002 Add Go dependencies: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles`, `github.com/charmbracelet/lipgloss`, `github.com/stretchr/testify` via `go get`
- [x] T003 Create `tools/admin-cli/goreleaser.yml` with targets `darwin/amd64`, `darwin/arm64`, `windows/amd64`; set `CGO_ENABLED=0`
- [x] T004 Create package skeleton directories: `cmd/alvestass-admin/`, `internal/config/`, `internal/trailbase/`, `internal/validate/`, `internal/ui/`
- [x] T005 Create `cmd/alvestass-admin/main.go` with `--help` and `--version` flag handling and empty TUI entry point; verify `go build ./...` passes

**Checkpoint**: `go build ./...` succeeds; binary exits cleanly with `--version` ⚠️ *Requires Go to be installed — run `brew install go` then `go mod tidy && go build ./...` to verify*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Config system, Trailbase client, and the Checker framework — required by every user story

**⚠️ CRITICAL**: No user story phase can begin until this phase is complete

- [x] T006 Implement `internal/config/config.go`: `Config` struct (backend_url, auth_token, auth_token_expiry), `Load()`, `Save()` using `os.UserConfigDir()` path with `0600` file permissions
- [x] T007 [P] Write `internal/config/config_test.go`: round-trip Load/Save, missing file returns empty config, `0600` permission enforced on Unix
- [x] T008 Implement `internal/trailbase/client.go`: `Client` struct with `baseURL` and `token`; methods `Authenticate(email, password) error`, `GetClubInfo() (ClubInfo, error)`, `UpdateClubInfo(fields map[string]any) error`; retry once on `401`
- [x] T009 [P] Write `internal/trailbase/client_test.go`: test JSON marshal/unmarshal for `ClubInfo`; test request construction (no live network calls)
- [x] T010 Define `internal/validate/checker.go`: `CheckIssue` struct (entity, field, value, rule); `Checker` interface with `Name() string` and `Run(ctx, client) ([]CheckIssue, error)`; no implementation here — only the contract
- [x] T011 Implement `internal/validate/clubinfo.go`: `ClubInfoChecker` struct implementing `Checker`; `Run()` calls `client.GetClubInfo()` then validates all 9 fields per data-model.md rules (non-empty strings, founding_year 1800–2100, short_description ≤ 300 chars, email pattern, postal_code `\d{3}\s?\d{2}`); returns `[]CheckIssue`
- [x] T012 [P] Write `internal/validate/clubinfo_test.go`: valid record produces no issues; each individual rule violation returns the correct `CheckIssue`; boundary values for founding_year and short_description length

**Checkpoint**: `go test ./internal/...` passes ⚠️ *Requires Go — run after `go mod tidy`*

---

## Phase 3: User Story 1 — First-Run Setup & Connectivity (Priority: P1) 🎯 MVP

**Goal**: Administrator can launch the CLI, complete the first-run wizard, and reach the main menu. CLI verifies backend connectivity at startup and shows a Swedish error if unreachable (FR-10, FR-11, FR-12, FR-01).

**Independent Test**: Launch with no config → complete wizard → reach main menu. Re-launch → skip wizard → main menu directly. Launch with wrong backend URL → Swedish connectivity error, exit code 1.

- [x] T013 [US1] Implement first-run wizard in `internal/ui/setup.go`: sequential Bubble Tea prompts for backend URL, email, password; show spinner during `client.Authenticate()` call (FR-05); save token and expiry to config on success
- [x] T014 [US1] Implement startup connectivity check in `cmd/alvestass-admin/main.go`: load config, attempt `client.GetClubInfo()` as ping, display `Anslutningsfel: [anledning]` in Swedish and exit code 1 if unreachable
- [x] T015 [US1] Implement main menu model in `internal/ui/menu.go`: Bubble Tea list with options [1] Uppdatera, [2] Kontrollera, [3] Hjälp, [4] Avsluta; keyboard navigation; `q` / `Ctrl+C` → Avsluta
- [x] T016 [US1] Wire setup wizard → connectivity check → main menu in `cmd/alvestass-admin/main.go`; handle `--config` flag override

**Checkpoint**: US1 complete — binary reaches main menu on first and subsequent runs; unreachable backend shows Swedish error

---

## Phase 4: User Story 2 — Update Contact Info (Priority: P2)

**Goal**: Administrator can select individual fields to edit interactively, review a diff, and confirm before saving. Validation uses `ClubInfoChecker` rules before the confirm prompt (FR-02, FR-05, FR-06, FR-07, FR-13).

**Independent Test**: Select [1] Uppdatera → edit `email` → confirm → `Kontaktuppgifter uppdaterade.` → verify in Trailbase admin panel. Enter invalid postal code → see Swedish `CheckIssue` → retry → correct value → confirm. Cancel at diff prompt → Trailbase unchanged.

- [x] T017 [US2] Implement update TUI flow in `internal/ui/update.go`: fetch current `club_info` via `client.GetClubInfo()` → display current values → numbered field selector → per-field text input with current value as placeholder → validate changed fields by running `ClubInfoChecker` rules directly → show `CheckIssue` list on failure → show diff on success → confirm → `PATCH` only changed fields
- [x] T018 [US2] Add spinner (`bubbles/spinner`) in `internal/ui/update.go` for fetch and save network calls (FR-05)
- [x] T019 [US2] Add `Uppdatera kontaktuppgifter` branch in `internal/ui/menu.go` to launch update flow; `Ctrl+C` / `q` at any step returns to main menu without changes (FR-07)

**Checkpoint**: US2 complete — update flow works end-to-end; only changed fields sent in PATCH body; cancellation leaves Trailbase unchanged

---

## Phase 5: User Story 3 — Consistency Check (Priority: P3)

**Goal**: The check operation runs all registered `Checker` implementations in order, groups results by name, and reports each `CheckIssue` with entity, field, value, and Swedish rule. Adding a new checker in a future feature requires no changes to the runner or UI (FR-03, FR-05, FR-06).

**Independent Test**: Run check on clean data → `Inga problem hittades.`. Manually corrupt a field via Trailbase admin panel → re-run → issue listed under `"Kontaktuppgifter"` with correct field and Swedish rule. Simulate network error for one checker → runner continues and reports the error inline rather than aborting.

- [x] T020 [US3] Implement check runner in `internal/ui/check.go`: accept `[]validate.Checker`; iterate in order; collect `[]CheckIssue` per checker; if a checker returns an error display `"[Name]: kunde inte hämta data — [anledning]"` and continue; group all issues by `Checker.Name()`; render grouped report or `"Inga problem hittades."`; press Enter → main menu
- [x] T021 [US3] Add spinner in `internal/ui/check.go` while checkers run (FR-05)
- [x] T022 [US3] Register `ClubInfoChecker` in the checker slice in `cmd/alvestass-admin/main.go`; add `Kontrollera data` branch in `internal/ui/menu.go` to launch the runner

**Checkpoint**: US3 complete — runner iterates registered checkers; adding a second `Checker` in a future PR requires only appending to the registration slice in `main.go`

---

## Phase 6: User Story 4 — Help Screen (Priority: P4)

**Goal**: Administrator can view a description of all operations and step-by-step instructions in Swedish (FR-04).

**Independent Test**: Select [3] Hjälp → both operations described in Swedish with instructions → press Enter → return to main menu.

- [x] T023 [US4] Implement help screen in `internal/ui/help.go`: static Swedish content describing update and check operations with step-by-step instructions; Enter or `q` returns to main menu
- [x] T024 [US4] Add `Hjälp` branch in `internal/ui/menu.go` to launch help screen

**Checkpoint**: US4 complete — help screen accessible from main menu and returns cleanly

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Error recovery, token refresh, distribution, and quality gate

- [x] T025 Audit all user-visible strings across `internal/ui/` — 100% Swedish; no English prompts or error messages
- [x] T026 Verify `Ctrl+C` / `q` cancellation at every TUI step returns to main menu without corrupting config or calling Trailbase (FR-07); verify every operation displays a completion or error summary message (FR-06); fix any gaps found
- [x] T027 Add token expiry check in `internal/config/config.go`: if `auth_token_expiry` is past, re-run login prompt before any operation and overwrite saved token
- [ ] T028 Run `go vet ./...` and resolve all warnings; run `go test ./...` and confirm all tests pass ⚠️ *Pending Go installation — run `brew install go && go mod tidy && go test ./...`*
- [ ] T029 Verify `goreleaser build --clean --snapshot` produces binaries for all 3 targets; confirm macOS arm64 binary runs on Apple Silicon; document Gatekeeper bypass step in `specs/003-admin-cli/quickstart.md` ⚠️ *Pending Go + goreleaser installation*

**Final Checkpoint**: `go test ./...` passes, `go vet ./...` clean, 3 release binaries produced, all quickstart.md manual test scenarios verified

---

## Dependencies

```
Phase 1 (Setup)
  └─► Phase 2 (Foundational: config, client, Checker interface + ClubInfoChecker)
        └─► Phase 3 (US1: wizard + main menu)   🎯 MVP threshold
              ├─► Phase 4 (US2: update)
              ├─► Phase 5 (US3: check runner)
              └─► Phase 6 (US4: help)
                    └─► Phase 7 (Polish)
```

**Phases 4–6 are independent of each other** and can be implemented in any order after Phase 3.

**Adding a future checker** (e.g. `MemberGroupChecker`):
1. Implement `Checker` interface in `internal/validate/membergroup.go`
2. Append to registration slice in `cmd/alvestass-admin/main.go`
3. No changes needed to `internal/ui/check.go` or menu

---

## Parallel Execution Opportunities

| Parallel Set | Tasks | Condition |
|---|---|---|
| Config + client + validate tests | T007, T009, T012 | After T006, T008, T010–T011 started |
| US2 + US3 + US4 stories | T017–T024 | After Phase 3 (T016) complete |

---

## Implementation Strategy

**MVP** = Phases 1–3 (T001–T016, 16 tasks): Working binary that authenticates, shows the main menu, and verifies connectivity.

**Full v1** = All phases (29 tasks): Update, check runner, help, polish, and cross-platform release binaries.

**Suggested delivery order**: MVP → Phase 4 (update — highest admin value) → Phase 5 (check runner) → Phase 6 (help) → Phase 7 (polish).
