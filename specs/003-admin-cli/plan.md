# Implementation Plan: Admin CLI

**Branch**: `003-admin-cli` | **Date**: 2026-04-22 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/003-admin-cli/spec.md`

## Summary

A standalone Go CLI executable that lets administrators update and validate the club's contact information in Trailbase without requiring database access or developer involvement. The `club_info` record is always present (seeded by migration); only interactive update and consistency check operations are needed. Distributed as a single binary for macOS (arm64/amd64) and Windows (amd64); uses Bubble Tea for the menu-driven TUI and Trailbase's REST API as the sole backend.

---

## Technical Context

**Language/Version**: Go 1.23+  
**Primary Dependencies**: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles`, `github.com/charmbracelet/lipgloss`, `github.com/stretchr/testify` (tests only)  
**Storage**: Local config file at `$UserConfigDir/alvestass-admin/config.json` (permissions `0600`); no local database  
**Testing**: Go standard `testing` package + `testify` for assertions  
**Target Platform**: macOS 14+ (arm64, amd64), Windows 10/11 (amd64)  
**Project Type**: CLI tool  
**Performance Goals**: Single PATCH completes within 5 s on a normal internet connection  
**Constraints**: Single binary, no runtime dependencies, `CGO_ENABLED=0`; credentials not stored in repo  
**Scale/Scope**: Single admin user; one Trailbase table (`club_info`, singleton) in v1

---

## Constitution Check

| Principle | Gate | Status | Notes |
|-----------|------|--------|-------|
| I — Code Quality | `go vet ./...` + `staticcheck` pass | ✓ No gate triggered | Spirit applies; Go type system replaces TypeScript strict mode |
| II — Testing | Unit tests for field validation rules (`internal/validate/`); `go build ./...` must pass | ✓ No gate triggered | Validation logic tested; integration via manual quickstart procedure |
| III — UX Consistency | Swedish text for all user-visible output | ✓ No gate triggered | CLI is not the website; Bulma/Alpine rules N/A |
| IV — Performance | 500 records imported in < 60 s | ✓ No gate triggered | Own success criterion defined in spec |
| V — Backend Architecture | Trailbase is sole backend | ✓ No gate triggered | CLI is a pure Trailbase REST client |
| VI — Idrottsarenan | N/A | ✓ Not applicable | No Idrottsarenan integration in this feature |
| VII — GDPR | `club_info` is organizational data; JWT stored at `0600` | ✓ No gate triggered | Migration explicitly flags `club_info` as non-personal. Future entity additions MUST re-evaluate. |

**No constitution violations. No blocking gates.**

---

## Project Structure

### Documentation (this feature)

```text
specs/003-admin-cli/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── cli-commands.md  # Phase 1 output — CLI invocation & menu contract
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code

```text
tools/admin-cli/
├── go.mod                          # module github.com/alvestass/admin-cli
├── go.sum
├── goreleaser.yml                  # Cross-compilation config (darwin/amd64, darwin/arm64, windows/amd64)
├── cmd/
│   └── alvestass-admin/
│       └── main.go                 # Entry point: parse flags, launch TUI
├── internal/
│   ├── config/
│   │   ├── config.go               # Load/save config.json; token expiry check
│   │   └── config_test.go
│   ├── trailbase/
│   │   ├── client.go               # HTTP client: auth, get, patch club_info
│   │   └── client_test.go          # Marshal/unmarshal tests (no network)
│   ├── validate/
│   │   ├── checker.go              # Checker interface + CheckIssue type
│   │   ├── clubinfo.go             # ClubInfoChecker implementation
│   │   └── clubinfo_test.go
│   └── ui/
│       ├── menu.go                 # Bubble Tea main menu model
│       ├── setup.go                # First-run wizard TUI flow
│       ├── update.go               # Update operation TUI flow
│       ├── check.go                # Consistency check TUI flow
│       └── help.go                 # Help screen
└── dist/                           # goreleaser output (git-ignored)
```

**Structure Decision**: Single Go module under `tools/admin-cli/`. Completely isolated from the Astro site (`src/`, `pnpm build`). Each concern separated into its own `internal/` package to allow independent testing.

---

## Complexity Tracking

No constitution violations requiring justification.
