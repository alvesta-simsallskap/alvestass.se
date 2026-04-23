# Data Model: Admin CLI

**Phase**: 1  
**Date**: 2026-04-22  
**Feature**: [spec.md](spec.md)

---

## Entities

### ClubInfo (backend — Trailbase `club_info` table)

Organizational contact information for the club. **Not personal data** (per migration comment — legitimate interest basis).

| Field | Type | Required | Constraints | Notes |
|-------|------|----------|-------------|-------|
| `id` | integer | yes | PK, always `1` (singleton) | Read-only |
| `name` | string | yes | non-empty | Club display name |
| `tagline` | string | yes | default `""` | Short motto |
| `founding_year` | integer | yes | 1800–2100 | `1921` for Alvesta SS |
| `short_description` | string | yes | max 300 chars | Displayed on homepage |
| `address` | string | yes | non-empty | Street address |
| `city` | string | yes | non-empty | |
| `postal_code` | string | yes | non-empty | Swedish format `NNN NN` |
| `phone` | string | yes | non-empty | |
| `email` | string | yes | non-empty, valid email format | |

**Singleton**: The table holds exactly one row (`id = 1`). Update operations always target this row via `PATCH /api/records/v1/club_info/1`.

**Validation rules** (enforced by CLI before any API call):
- All required fields must be non-empty strings after trimming whitespace
- `founding_year` must parse as integer in range [1800, 2100]
- `short_description` must be ≤ 300 characters
- `email` must match a basic email pattern (`x@y.z`)
- `postal_code` must match `\d{3}\s?\d{2}` (Swedish format)

---

### Config (local — `$UserConfigDir/alvestass-admin/config.json`)

Per-user configuration stored on the administrator's workstation.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `backend_url` | string | yes | e.g. `https://alvestass-trailbase.fly.dev` |
| `auth_token` | string | yes | Trailbase JWT (obtained via login) |
| `auth_token_expiry` | RFC3339 string | yes | CLI re-authenticates when expired |

**Security**:
- File created with permissions `0600` (Unix) on first write
- Email and password are NEVER stored; only the resulting JWT is persisted
- On token expiry, CLI prompts for credentials, obtains a new token, and overwrites the file

---

### CheckIssue (transient — produced by any Checker)

Describes a single data quality problem found by a health check. Replaces the earlier `ValidationError` type, which had a CSV-specific `row` field. `entity` makes issues self-describing regardless of which table they come from.

| Field | Type | Notes |
|-------|------|-------|
| `entity` | string | Record identifier, e.g. `"club_info/1"`, `"member/42"` |
| `field` | string | Field name within that record, e.g. `"email"` |
| `value` | string | The offending value (redact if PII — see GDPR note below) |
| `rule` | string | Human-readable rule violated, in Swedish |

**GDPR note**: Future checkers that operate on personal data (member names, emails) MUST redact `value` in any log output. Display in the CLI terminal is acceptable since it is an authenticated admin session, but the value must not be written to any file or external system.

---

### Checker (interface — `internal/validate/checker.go`)

The extensibility contract for the consistency check operation. Every health check — present and future — implements this interface. The check runner in `internal/ui/check.go` iterates a registered slice of `Checker` implementations and groups their results by name.

```
Checker
  Name() string                                              → display name shown in results, e.g. "Kontaktuppgifter"
  Run(ctx, client) ([]CheckIssue, error)                    → fetch data, validate, return issues
```

**Current implementations** (v1):
- `ClubInfoChecker` — fetches `club_info/1`, runs all field validation rules

**Planned future implementations** (not in scope for v1):
- `MemberGroupChecker` — verifies every member belongs to a training group or the board
- `InstructorTimeReportChecker` — verifies time reports exist for all instructors who worked in the active month

**Runner contract**: if a `Checker` returns a non-nil error (e.g. network failure), the runner reports it as `"[CheckName]: kunde inte hämta data — [anledning]"` and continues with remaining checkers. A failing check must not abort the whole operation.

---

## State Transitions

### CLI Session

```
Launch
  └─► Connectivity check
        ├─ FAIL ──► Show error + exit
        └─ OK ───► Main Menu
                      ├─► Update ──► Show current ──► Field prompts ──► Validate ──► Confirm ──► Apply ──► Summary ──► Main Menu
                      ├─► Check  ──► Run all Checkers ──► Group results by name ──► Report ──► Main Menu
                      ├─► Help   ──► Help screen ──► Main Menu
                      └─► Exit   ──► Goodbye message + exit
```

### Token Lifecycle

```
Load config
  ├─ No config ──► First-run wizard ──► Login ──► Save token ──► Continue
  ├─ Token expired ──► Re-login prompt ──► Login ──► Save token ──► Continue
  └─ Token valid ──► Continue
```

---

## Data Flow: Check Operation (runner pattern)

```
Admin selects "Kontrollera data"
  └─► For each registered Checker (in order):
        ├─ Run(ctx, client)
        │     ├─ Error ──► Record "[Name]: kunde inte hämta data — [anledning]"
        │     └─ OK    ──► Collect []CheckIssue
        └─► Group all issues by Checker.Name()
              └─► Display grouped report
                    ├─ No issues ──► "Inga problem hittades."
                    └─ Issues    ──► List per group → press Enter → Main Menu
```

---

## Data Flow: Update Operation

```
Admin selects "Uppdatera kontaktuppgifter"
  └─► GET /api/records/v1/club_info/1
        └─► Display current field values
              └─► Admin selects field(s) to edit → enters new values
                    └─► Validate all changed fields (ClubInfoChecker rules)
                          ├─ Errors ──► Display CheckIssue list ──► Retry or cancel
                          └─ Valid ──► Show diff ──► Prompt "Spara? (j/n)"
                                            ├─ n ──► Abort → Main Menu
                                            └─ j ──► PATCH /api/records/v1/club_info/1
                                                          └─► "Kontaktuppgifter uppdaterade." → Main Menu
```
