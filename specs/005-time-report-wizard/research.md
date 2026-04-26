# Research: Two-Step Time Report Wizard

**Feature**: 005-time-report-wizard  
**Date**: 2026-04-26

---

## Decision 1: SQLite STRICT table schema migration (swim_school_rate nullability)

**Problem**: SQLite does not support `ALTER COLUMN`. The existing `instructors`
table has `swim_school_rate INTEGER NOT NULL`. We need to make it nullable.

**Decision**: Use the SQLite table-rename recreation pattern:
1. Create `instructors_new` with the desired schema.
2. Copy all rows from `instructors` into `instructors_new`.
3. Drop `instructors`.
4. Rename `instructors_new` to `instructors`.

All four steps run inside the same Trailbase migration file. Trailbase wraps
each migration in a transaction, so a failure mid-migration rolls back cleanly.

**Rationale**: This is SQLite's documented procedure for schema changes that
`ALTER TABLE` cannot express. It is safe inside a transaction and requires no
data transformation (existing `swim_school_rate` values are always > 0, which
satisfies the new nullable `CHECK`).

**Alternatives considered**:
- `ALTER TABLE instructors ADD COLUMN travel_compensation` then leave
  `swim_school_rate` as-is. Rejected: does not satisfy FR-013 (coach-only
  instructors need `swim_school_rate` NULL).

---

## Decision 2: New CHECK constraint — at least one rate required

**Decision**: Add `CHECK(swim_school_rate IS NOT NULL OR coach_rate IS NOT NULL)`
to the recreated `instructors` table.

**Rationale**: FR-013 states "At least one rate MUST be set for a valid
instructor profile." Enforcing this at the database layer prevents admin UI
mistakes from creating orphaned instructor records that would cause a confusing
"no sections shown" state in the wizard.

**Alternatives considered**:
- Application-level validation only (in the lookup endpoint). Rejected: the
  database is the single source of truth; application checks alone are
  insufficient.

---

## Decision 3: New `/api/lookup-instructor` endpoint — server-side proxy pattern

**Decision**: Add a new SSR API route `src/pages/api/lookup-instructor.ts`
that accepts `POST { email }`, authenticates against Trailbase as the service
user, fetches the instructor record, and returns only
`{ swimSchool, coach, travelCompensation }`. Salary rates are never sent to
the browser.

**Rationale**:
- Data minimization (GDPR Art. 5(1)(c)): the browser only needs role flags and
  the travel compensation flag — no salary rates.
- Privacy: the Trailbase service user credentials stay on the server.
- Consistent with the existing pattern: all other Trailbase calls go through
  Cloudflare Worker SSR (`tidrapport.astro`, `send-time-report.ts`).

**Alternatives considered**:
- Client-side direct Trailbase call. Rejected: would require exposing the
  service user token to the browser, violating Principle V.
- Server-side render of the full form with email as a query parameter (redirect
  after step 1). Rejected: email in URL violates GDPR Principle VII ("no
  personal data in URLs").

---

## Decision 4: Alpine.js `x-if` (template) for conditional sections

**Decision**: Use Alpine.js `<template x-if="...">` (not `x-show`) for all
role-conditional sections in step 2 (Simskola, Träningsgrupper, Utlägg,
Milersättning).

**Rationale**: `x-if` removes the element from the DOM when false; `x-show`
only sets `display:none`. SC-003 explicitly requires the Milersättning field to
be absent from the DOM for ineligible instructors. Using `x-if` consistently
across all conditional sections simplifies the implementation and satisfies the
spec for all gated fields.

**Alternatives considered**:
- `x-show` for section boxes, `x-if` only for Milersättning. Rejected: mixing
  two conditional mechanisms for the same pattern adds complexity without
  benefit.

---

## Decision 5: Alpine.js state machine for two-step flow

**Decision**: The `tidrapport.astro` page wraps the entire form in a single
top-level Alpine.js `x-data` component with state:

```js
{
  step: 1,         // 1 = email input, 2 = report form
  email: '',       // bound to the step 1 email field
  loading: false,  // true while lookup is in flight
  error: null,     // Swedish error string | null
  role: null,      // { swimSchool, coach, travelCompensation } | null
}
```

Step 1 submits via an Alpine method (`lookupEmail`) that POSTs to
`/api/lookup-instructor`, stores the role flags in `this.role`, and sets
`this.step = 2`. Step 2 reads `this.role` to show/hide sections. The hidden
`<input name="email">` in the step 2 form is bound to `this.email` so it is
included in the final form submission.

**Rationale**: Single Alpine component avoids cross-component communication.
The pattern matches how the existing `formHandler` data component works.
No page reload between steps means the already-loaded schedule data (fetched
at SSR time) is reused without a second server round-trip.

**Alternatives considered**:
- Two separate pages (step 1 redirects to step 2 with a session cookie). 
  Rejected: adds complexity, requires a cookie-writing API route, and email in
  cookie is a new personal-data store.
- Two Astro components each with their own `x-data`. Rejected: role state would
  need to be passed between components via events or global state; more complex
  with no benefit.
