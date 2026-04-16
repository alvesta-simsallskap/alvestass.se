# Research: Time Report Trailbase Migration

**Feature**: 002-time-report-trailbase
**Date**: 2026-04-16
**Status**: Complete

---

## Decision 1: Service-to-Service Authentication (Employee Lookup)

**Decision**: The Cloudflare Worker authenticates to Trailbase on each form
submission using a dedicated service user account. Credentials (email + password)
are stored as Cloudflare Worker secrets. The Worker POSTs to
`/api/auth/v1/token`, receives a short-lived JWT, and immediately uses it to
fetch the employee record.

**Rationale**: Trailbase v0.26.3 does not support long-lived API tokens or
service accounts (planned for a future release). Re-authenticating on each
submission is the correct approach for now:
- Form submissions are low-frequency (a few per month per instructor)
- The extra round-trip latency (~50–100 ms to arn) is negligible for a form
- No token caching infrastructure (KV, etc.) is needed
- Credentials are never in source code

**Alternatives considered**:
- Cache JWT in Cloudflare KV — rejected (adds KV dependency and complexity for
  negligible benefit given submission frequency).
- Store a pre-generated admin JWT as a Worker secret — rejected (1-hour TTL
  makes it impractical; using admin credentials for service calls is poor
  practice).
- Make instructors table public — rejected (violates GDPR; employee emails and
  salary rates are personal data and MUST NOT be served via unauthenticated
  API routes, per constitution Principle VII).

**New secrets required**:
- `TRAILBASE_SERVICE_EMAIL` — email of the dedicated Trailbase service user
- `TRAILBASE_SERVICE_PASSWORD` — password of the dedicated Trailbase service user

---

## Decision 2: Trailbase Records API Filter Syntax

**Decision**: Use the `filter[column][$eq]=value` query parameter syntax for
all Trailbase record lookups. Multiple filters are AND'd automatically.

**Verified filter pattern** (multi-field):
```
GET /api/records/v1/time_report_sessions
  ?filter[month_key][$eq]=2026-04
  &filter[group][$eq]=simskola
  &limit=500
```

**Single-field lookup** (employee by email):
```
GET /api/records/v1/instructors
  ?filter[email][$eq]=coach@example.com
  &limit=1
  Authorization: Bearer <jwt>
```

**Rationale**: The filter syntax is documented and stable. A `limit=500` on
session fetches is safe — a month typically has fewer than 100 sessions across
all groups. Fetching all sessions for a month in a single request (then
grouping client-side in the Worker) is preferable to 6 separate per-group
requests.

---

## Decision 3: /tidrapport Page Rendering Mode

**Decision**: Make `/tidrapport` SSR (`export const prerender = false`) with
a 1-hour edge cache:
```
Cache-Control: public, max-age=3600, stale-if-error=604800
```

**Rationale**: The session schedule and configuration must be loaded from
Trailbase at render time — they cannot be baked into the static build. SSR is
required. A 1-hour cache (SC-001 requires only "within 5 minutes", which is
the config/employee case; the form schedule only changes monthly) balances
freshness against the fly.io free-tier cold-start cost. `stale-if-error` of
7 days ensures the form remains usable if Trailbase is unreachable.

**Fly.io cold-start consideration**: The `auto_stop_machines = 'stop'` setting
in fly.toml means a cold start (5–10 s) may occur after inactivity. The 1-hour
edge cache mitigates this for returning visitors. For the first visitor after
a cold start, the form may take longer to load — this is acceptable for an
internal tool.

**Alternative considered**: Client-side fetch via Alpine.js — rejected (adds
client-side JS bundle, contradicts Principle IV; also requires CORS
configuration on Trailbase).

---

## Decision 4: `IS_DEVELOPMENT` Flag Replacement

**Decision**: Replace the hardcoded `IS_DEVELOPMENT = false` in
`time-report-settings.ts` with Astro's built-in `import.meta.env.DEV`.
This boolean is `true` during `pnpm dev` and `false` in production builds.

**Rationale**: The current implementation requires a manual code edit each
time a developer wants to test locally. `import.meta.env.DEV` is the standard
Astro/Vite pattern and requires no configuration change.

---

## Decision 5: Data Table Access Policies

**Decision** *(amended 2026-04-16 — see note below)*:

| Table | Read | Write |
|-------|------|-------|
| `time_report_config` | Authenticated | Admin only |
| `time_report_sessions` | Authenticated | Admin only |
| `instructors` | Authenticated | Admin only |

**Amendment note**: Originally `time_report_config` and `time_report_sessions`
were planned as "Everyone" (public read). This was revised after clarification
that **all backend API endpoints MUST require the service user token** — no
endpoint is publicly accessible (FR-008, spec session 2026-04-16). Consequently:
- The service user authenticates at the start of **every Worker invocation**
  (page load and form submission), not only at form submission time.
- The single login token is reused for all Trailbase calls within that request.
- Added latency: ~80–150 ms per page load for the fly.io login roundtrip.
  This is acceptable given the schedule-caching approach (Decision 3).

**Original rationale** (still valid for instructors/instructors):
- `instructors` contains email addresses and salary rates (personal data under
  GDPR). Authenticated-only read access satisfies Principle VII.

---

## Decision 6: Session Data Fetch Strategy

**Decision**: Fetch all sessions for the active month in a single request
(`?filter[month_key][$eq]=<key>&limit=500`). Group them by group in the
Worker/Astro frontmatter before passing to the template.

**Rationale**: Simpler than 6 group-specific requests. Typical months have
50–80 sessions total, well within the 500 limit. Grouping logic is trivial.

---

## Decision 7: `time-report-items.json` Removal

**Decision**: Delete `src/config/time-report-items.json` once the migration
is complete and all months needed for the upcoming season are entered into
Trailbase.

**Rationale**: The JSON file is the data source the Trailbase `time_report_sessions`
table replaces. Historical data (prior months) does not need to be migrated
since the form is only used for the active month. The current month's sessions
will be entered into Trailbase by the admin as part of the migration.

---

## GDPR Notes

- **instructors table**: Personal data (email, salary rates). Legal basis:
  contractual necessity (employment relationship). Retention: until termination
  of employment + 1 year for accounting. Must be added to `docs/gdpr-register.md`
  (TODO: GDPR_REGISTER) before go-live.
- **time_report_config** and **time_report_sessions**: No personal data.
- **Form submissions**: No personal data stored in Trailbase (FR-007). Email
  continues to be the sole record. No new GDPR storage obligations.
- **TODO(INTEGRITETSPOLICY)**: The `/integritetspolicy` page must exist before
  any feature storing personal data goes live. The instructors table introduction
  constitutes the first personal-data storage in Trailbase — this TODO must be
  resolved before go-live.
