# API Contract: POST /api/lookup-instructor

**Feature**: 005-time-report-wizard  
**File**: `src/pages/api/lookup-instructor.ts`  
**Method**: POST  
**Auth**: None (public endpoint — returns no personal data beyond role flags)  
**SSR**: Yes (`export const prerender = false`)

---

## Purpose

Resolves an instructor's identity from their email address and returns the
role flags needed to render the step 2 form. Called client-side by Alpine.js
when the instructor submits step 1.

---

## Request

```
POST /api/lookup-instructor
Content-Type: application/json

{
  "email": "instructor@example.com"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `email` | string | Yes | Non-empty; basic format check (contains `@`) |

---

## Responses

### 200 OK — Instructor found

```json
{
  "swimSchool": true,
  "coach": false,
  "travelCompensation": false
}
```

| Field | Type | Meaning |
|-------|------|---------|
| `swimSchool` | boolean | `swim_school_rate IS NOT NULL` — show Simskola section |
| `coach` | boolean | `coach_rate IS NOT NULL` — show Träningsgrupper + Utlägg sections |
| `travelCompensation` | boolean | `travel_compensation = 1` — show Milersättning field |

### 404 Not Found — Email not registered

```json
{ "error": "not_found" }
```

### 400 Bad Request — Invalid email format

```json
{ "error": "invalid_email" }
```

### 503 Service Unavailable — Trailbase unreachable

```json
{ "error": "backend_unavailable" }
```

---

## Privacy & Security

- **No salary rates are returned** — the response contains only boolean flags.
- **Email MUST NOT be logged** — do not include the email address in any
  `console.log`, `console.error`, or Worker log output.
- **Brute-force note**: this endpoint reveals whether an email is registered.
  For an internal club tool with ~30 known instructors this is acceptable.
  No rate limiting is required at this stage.

---

## Implementation Notes

- Authenticate as service user (same `TRAILBASE_SERVICE_EMAIL` /
  `TRAILBASE_SERVICE_PASSWORD` secrets used by `tidrapport.astro`).
- Reuse the existing `authenticateServiceUser` and `fetchInstructor` functions
  from `src/lib/trailbase.ts`.
- If Trailbase returns non-2xx, return 503.
- If Trailbase returns 0 records, return 404.
- Map the `Instructor` record to `InstructorRole` before responding:
  ```
  swimSchool       = instructor.swim_school_rate !== null
  coach            = instructor.coach_rate !== null
  travelCompensation = instructor.travel_compensation === true
  ```
- Set `Cache-Control: no-store` on the response.
