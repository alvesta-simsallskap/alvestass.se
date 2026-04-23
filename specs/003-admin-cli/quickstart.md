# Quickstart: Admin CLI

**Date**: 2026-04-22  
**Feature**: [spec.md](spec.md)

---

## Prerequisites

- Go 1.23+ installed (`go version`)
- `goreleaser` installed (optional for local builds — `brew install goreleaser` or `go install github.com/goreleaser/goreleaser/v2@latest`)
- Trailbase instance running and accessible (see `TRAILBASE_URL` in `.dev.vars`)
- An admin account on the Trailbase instance

---

## Local Development Build

```bash
cd tools/admin-cli
go build -o alvestass-admin ./cmd/alvestass-admin
./alvestass-admin
```

The first run triggers the config wizard. Your config is saved to:
- macOS: `~/Library/Application Support/alvestass-admin/config.json`
- Windows: `%APPDATA%\alvestass-admin\config.json`

---

## Running Tests

```bash
cd tools/admin-cli
go test ./...
```

Tests cover CSV parsing and field validation. No network calls in the test suite.

---

## Cross-Compilation (release builds)

```bash
cd tools/admin-cli
goreleaser build --clean --snapshot
```

Produces binaries under `dist/`:
- `alvestass-admin_darwin_amd64/alvestass-admin` — macOS Intel
- `alvestass-admin_darwin_arm64/alvestass-admin` — macOS Apple Silicon
- `alvestass-admin_windows_amd64/alvestass-admin.exe` — Windows 64-bit

---

## Manual Test Procedure (golden path)

Run these steps against a real Trailbase instance before marking any PR as done.

### 1. First-run config wizard
```
./alvestass-admin
```
- Expect: prompted for backend URL, email, password
- Expect: "Ansluter... ✓" and config file created
- Verify: `cat ~/Library/Application\ Support/alvestass-admin/config.json` shows `backend_url` and `auth_token`

### 2. Import contact info
- Create a test CSV:
  ```csv
  name,tagline,founding_year,short_description,address,city,postal_code,phone,email
  Alvesta Simsällskap,Simglädje sedan 1921,1921,Simglädje.,Hjortsbergavägen 6C,Alvesta,342 36,076 027 94 10,kansli@alvestass.se
  ```
- Select [1] Import → provide file path → confirm
- Expect: `Importerade kontaktuppgifter. 1 rad importerad, 0 fel.`
- Verify: website `kontakt` page shows updated data

### 3. Update a field
- Select [2] Uppdatera → choose `tagline` → enter new value → confirm
- Expect: `Kontaktuppgifter uppdaterade.`
- Verify: change visible via Trailbase admin panel

### 4. Consistency check
- Select [3] Kontrollera data
- Expect: `Inga problem hittades.` (if data is clean)
- Manually corrupt a field in Trailbase admin → re-run check → expect violation listed

### 5. Backend unreachable
- Stop Trailbase (or use a wrong URL in config)
- Launch CLI
- Expect: Swedish error message about backend being unreachable; exit code 1

### 6. Cancel during import (FR-08)
- Start import → at confirm prompt type `n`
- Expect: returns to main menu; Trailbase data unchanged

---

## Distributing to an Administrator

1. Build the binary for the target platform (see Cross-Compilation above)
2. Share the binary via email, Dropbox, or the GitHub Releases page
3. macOS: right-click → Open (first time) to bypass Gatekeeper, or `xattr -d com.apple.quarantine ./alvestass-admin`
4. Windows: double-click `alvestass-admin.exe` or run from Command Prompt

No installation steps needed beyond placing the file somewhere convenient.
