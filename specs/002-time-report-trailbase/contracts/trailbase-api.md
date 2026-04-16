# API Contract: Time Report Trailbase Endpoints

**Feature**: 002-time-report-trailbase
**Date**: 2026-04-16

All endpoints are on the Trailbase instance at `https://alvestass-trailbase.fly.dev`.

---

## 1. Fetch Active Configuration

**Used by**: `tidrapport.astro` SSR frontmatter

```
GET /api/records/v1/time_report_config?limit=1
Authorization: Bearer <auth_token>
```

**Response**:
```json
{
  "records": [
    {
      "id": 1,
      "active_month_key": "2026-04",
      "active_month_display": "april 2026",
      "extra_time_simskola": 30,
      "extra_time_training": 15,
      "half_day_salary": 500,
      "full_day_salary": 1000,
      "overnight_salary": 300
    }
  ],
  "cursor": null
}
```

**Error handling**: If the response has no records or the request fails, the
page must render a Swedish error message (same pattern as `kontakt.astro`).

---

## 2. Fetch Sessions for Active Month

**Used by**: `tidrapport.astro` SSR frontmatter

```
GET /api/records/v1/time_report_sessions
    ?filter[month_key][$eq]=<active_month_key>
    &limit=500
Authorization: Bearer <auth_token>
```

**Response**:
```json
{
  "records": [
    {
      "id": 42,
      "month_key": "2026-04",
      "training_group": "tavlingA",
      "date": "2026-04-07",
      "title": "Träning",
      "hours": 1,
      "minutes": 45
    }
  ],
  "cursor": null
}
```

**Grouping**: The fetched array is grouped by `group` into a `SessionsByGroup`
object before being passed to the template.

**Error handling**: If sessions cannot be fetched, the page renders each group
as empty. The form remains usable for extra-time and expense entries.

---

## 3. Authenticate as Service User

**Used by**: Both `tidrapport.astro` SSR (page load) and `send-time-report.ts` (form submission) — called once per Worker invocation; the token is reused for all subsequent Trailbase calls within that invocation

```
POST /api/auth/v1/token
Content-Type: application/json

{
  "email": "<TRAILBASE_SERVICE_EMAIL>",
  "password": "<TRAILBASE_SERVICE_PASSWORD>"
}
```

**Response**:
```json
{
  "auth_token": "<jwt>",
  "refresh_token": "<refresh_jwt>",
  "csrf_token": "..."
}
```

Only `auth_token` is used. The token is short-lived (~1 hour) and is used
immediately in the same request handler — no caching or refresh is needed.

**Error handling**: If authentication fails, the form submission continues
without a salary estimate. The report is still emailed. A server-side log
entry is written (but no personal data is logged).

---

## 4. Fetch Instructor by Email

**Used by**: `send-time-report.ts` after successful auth

```
GET /api/records/v1/instructors
    ?filter[email][$eq]=<submitter_email>
    &limit=1
Authorization: Bearer <auth_token_from_step_3>
```

**Response (match found)**:
```json
{
  "records": [
    {
      "id": 7,
      "email": "coach@example.com",
      "swim_school_rate": 115,
      "coach_rate": 145
    }
  ],
  "cursor": null
}
```

**Response (no match)**:
```json
{
  "records": [],
  "cursor": null
}
```

**Error handling**: If no instructor is found for the email, the salary estimate
is omitted from the email (existing behaviour for unknown emails). If the
Trailbase call fails, the salary estimate is omitted and the report is still
sent.

---

## TypeScript Client Functions (to add to `src/lib/trailbase.ts`)

```typescript
// Authenticate as the service user and return the JWT.
// Called once at the start of each Worker invocation.
export async function authenticateServiceUser(
  baseUrl: string,
  email: string,
  password: string
): Promise<string>  // returns auth_token

// Fetch the single config row (requires auth_token)
export async function fetchTimeReportConfig(
  baseUrl: string,
  authToken: string
): Promise<TimeReportConfig | null>

// Fetch all sessions for a given month key (requires auth_token)
export async function fetchTimeReportSessions(
  baseUrl: string,
  monthKey: string,
  authToken: string
): Promise<TimeReportSession[]>

// Fetch a single instructor by email (requires auth_token)
export async function fetchInstructor(
  baseUrl: string,
  email: string,
  authToken: string
): Promise<Instructor | null>
```
