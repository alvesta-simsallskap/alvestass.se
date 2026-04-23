# Research: Admin CLI

**Phase**: 0  
**Date**: 2026-04-22  
**Feature**: [spec.md](spec.md)

---

## Decision 1: TUI library

**Decision**: `github.com/charmbracelet/bubbletea` (Bubble Tea) + `github.com/charmbracelet/bubbles` + `github.com/charmbracelet/lipgloss`

**Rationale**: The Elm-architecture model fits the spec's requirement for a menu-driven interface with clear state transitions (main menu → operation → confirm → result). `bubbles` ships ready-made components for text inputs, spinners, and list selectors — covering FR-01, FR-03, and FR-06 with minimal custom code. `lipgloss` provides terminal styling without requiring ANSI escape sequences by hand.

**Alternatives considered**:
- `github.com/AlecAivazis/survey/v2` — simpler prompt library, good for linear wizard flows but lacks the composable state machine needed for FR-08 (mid-operation cancel). Ruled out.
- Plain `fmt.Scan` / `bufio.Scanner` prompts — zero dependencies, but produces a brittle interactive loop with no spinner or structured list selection. Ruled out for UX quality.
- `github.com/rivo/tview` — full TUI framework (panels, tables). More power than needed; adds ~3 MB to binary. Ruled out.

---

## Decision 2: Cross-compilation and distribution

**Decision**: `github.com/goreleaser/goreleaser` for building and packaging; targets `darwin/amd64`, `darwin/arm64`, `windows/amd64`.

**Rationale**: GoReleaser automates cross-compilation with `CGO_ENABLED=0` (required for truly static binaries on both platforms), produces correctly named artefacts, and can publish to GitHub Releases. A single `goreleaser.yml` in the `tools/admin-cli/` directory covers all targets declared in FR-09 and FR-10.

**Alternatives considered**:
- Manual `GOOS=... go build` scripts — works but fragile; no checksum generation or release packaging. Ruled out for anything beyond one-off local builds.
- Nix cross-compilation — reproducible but complex to set up for contributors on Windows. Ruled out.

**CGO note**: All selected dependencies (`bubbletea`, `bubbles`, `lipgloss`) are pure Go. No CGO. `CGO_ENABLED=0` is safe and required for Windows cross-compilation from macOS.

---

## Decision 3: Configuration storage

**Decision**: `os.UserConfigDir()` as the base; store config at `$UserConfigDir/alvestass-admin/config.json`. File permissions set to `0600` (owner read/write only) on creation.

**Rationale**: `os.UserConfigDir()` resolves to:
- macOS: `~/Library/Application Support`
- Windows: `%APPDATA%`
- Linux: `$XDG_CONFIG_HOME` or `~/.config`

This satisfies FR-12 (per-user, not shared, not version-controlled). The `0600` permission covers FR-12's security requirement on Unix; Windows ACLs via normal file creation restrict to the creating user by default.

**Config structure**:
```json
{
  "backend_url": "https://alvestass-trailbase.fly.dev",
  "auth_token":  "<jwt>",
  "auth_token_expiry": "2026-05-22T00:00:00Z"
}
```

The CLI authenticates via Trailbase's `POST /api/auth/v1/login` on first run (or when token is expired) and caches the resulting JWT. The administrator provides email + password on first launch; these are NOT stored — only the resulting token is.

**Alternatives considered**:
- Storing raw email/password — simpler but violates the principle of minimum credential exposure. Ruled out.
- Environment variables only — no persistence; requires re-entry on every run. Ruled out.
- OS keychain (e.g., `go-keyring`) — ideal for secrets but adds a native dependency that breaks pure-Go cross-compilation. Ruled out for v1; document as a future improvement.

---

## Decision 4: HTTP client and Trailbase integration

**Decision**: Go standard library `net/http` + `encoding/json`. No third-party HTTP client.

**Rationale**: Trailbase exposes a straightforward REST API (observed in `src/lib/trailbase.ts`):
- Auth: `POST /api/auth/v1/login` → `{ "auth_token": "..." }`
- List: `GET /api/records/v1/{table}?limit=N`
- Get one: `GET /api/records/v1/{table}/{id}`
- Update: `PATCH /api/records/v1/{table}/{id}` with JSON body
- Create: `POST /api/records/v1/{table}` with JSON body

Standard library handles this cleanly. The only external dependency is Bubble Tea, so keeping the HTTP layer in stdlib minimizes binary size and dependency surface.

**Request pattern**: All authenticated requests include `Authorization: Bearer <token>`. On `401` response the CLI re-authenticates and retries once before reporting an error.

---

## Decision 5: CSV import format for contact info

**Decision**: Single-row CSV with a header row matching `club_info` column names. The import operation replaces the single `club_info` record (upsert by `id=1`).

**Rationale**: The `club_info` table has exactly one row (singleton configuration record). "Import" for this entity means: parse one CSV row, validate all required fields, then `PATCH /api/records/v1/club_info/1`. A multi-row CSV would not make sense given the schema.

**CSV columns** (header row, same order not required):
```
name,tagline,founding_year,short_description,address,city,postal_code,phone,email
```

**Alternatives considered**:
- JSON import — more expressive but harder for administrators to edit in Excel/Numbers. CSV preferred.
- YAML import — same argument against as JSON; also less familiar to non-technical users.

---

## Decision 6: Project location in the monorepo

**Decision**: New Go module at `tools/admin-cli/` in the repository root.

**Rationale**: The CLI is a development/admin tool, not part of the website runtime. Placing it under `tools/` signals its purpose clearly and keeps it isolated from `src/` (Astro) and `trailbase/` (backend service). The Go module (`go.mod`) declares `module github.com/alvestass/admin-cli`.

**Build isolation**: `pnpm build` (Astro) does not touch `tools/`. Go build (`go build ./...` inside `tools/admin-cli/`) does not touch `src/`. No cross-contamination.

---

## Decision 7: Testing approach

**Decision**: Go standard `testing` package + `github.com/stretchr/testify` for assertions. No mock HTTP server for v1 — validation logic is tested purely; Trailbase integration is tested manually.

**Rationale**: The logic worth unit-testing is CSV parsing + field validation (FR-15). This is pure functional code with no I/O. HTTP integration tests against a live Trailbase instance are deferred to the quickstart manual test procedure.

**Test files**:
- `internal/csv/parser_test.go` — CSV parsing and validation rules
- `internal/trailbase/client_test.go` — request construction (no network calls; tests marshal/unmarshal only)

---

## Constitution Compliance Summary

| Principle | Status | Notes |
|-----------|--------|-------|
| I — Code Quality | ✓ Adapted | Go strict mode via `go vet` + `staticcheck`. No `any` equivalent used. |
| II — Testing | ✓ | Unit tests for validation and CSV logic. Build gate: `go build ./...` |
| III — UX Consistency | N/A | CLI is not the website. Swedish text requirement DOES apply (FR per spec). |
| IV — Performance | N/A | Own success criterion: 500 records in 60 s. |
| V — Backend Architecture | ✓ | CLI is purely a Trailbase REST client. No competing database. |
| VI — Idrottsarenan | N/A | Not applicable to this feature. |
| VII — GDPR | ✓ Low risk | `club_info` is organizational data, explicitly non-personal (migration comment). JWT cached locally at `0600`. No PII in logs. Future entity additions MUST re-evaluate this gate. |
