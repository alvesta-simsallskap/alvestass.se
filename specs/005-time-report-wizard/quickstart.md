# Quickstart: Two-Step Time Report Wizard

**Feature**: 005-time-report-wizard  
**Date**: 2026-04-26

---

## Prerequisites

- Node.js + pnpm installed
- `.dev.vars` file with the required secrets (see CLAUDE.md)
- Trailbase instance running (local or staging)

---

## Running Locally

```bash
pnpm dev          # starts Astro dev server with Cloudflare Worker emulation
```

Open `http://localhost:4321/tidrapport` in a browser.

**Step 1 (email lookup)**: In dev mode the Turnstile widget is skipped. Enter
any email registered in the local Trailbase instance.

**Step 2 (form)**: Verify that:
- Simskola box appears only if the instructor has `swim_school_rate` set.
- Träningsgrupper box appears only if `coach_rate` is set.
- Övrig tid and Kommentarer appear for all instructors.
- Utlägg appears only if `coach_rate` is set.
- Milersättning appears only if `travel_compensation = 1`.

**Submitting**: In dev mode the email is replaced by an HTML preview in a new
tab (existing behavior).

---

## Applying the Database Migration

The migration file is:

```
trailbase/migrations/U1776686400__update_instructors.sql
```

To apply it against the local Trailbase instance, restart Trailbase — it runs
pending migrations automatically on start. Alternatively, run the SQL manually
in the Trailbase admin SQL editor.

After migration:
- Verify `instructors` table has `swim_school_rate` nullable and
  `travel_compensation` column present.
- Set `travel_compensation = 1` on a test instructor to verify the
  Milersättning field appears in step 2.
- Create a coach-only instructor (`swim_school_rate` NULL, `coach_rate` set)
  and verify only the coach sections appear.

---

## Testing the Email Lookup Endpoint Directly

```bash
curl -X POST http://localhost:4321/api/lookup-instructor \
  -H 'Content-Type: application/json' \
  -d '{"email":"coach@example.com"}'
```

Expected responses:
- Known coach-only email: `{"swimSchool":false,"coach":true,"travelCompensation":false}`
- Unknown email: `{"error":"not_found"}`
- Malformed body: `{"error":"invalid_email"}`

---

## Build Gate

```bash
pnpm build        # wrangler types && astro check && astro build
```

Must complete with zero TypeScript errors and zero `astro check` warnings.

---

## Files Changed

| File | Type |
|------|------|
| `trailbase/migrations/U1776686400__update_instructors.sql` | New |
| `src/lib/types.ts` | Updated |
| `src/lib/salary.ts` | Updated (null guard) |
| `src/pages/api/lookup-instructor.ts` | New |
| `src/pages/api/send-time-report.ts` | Updated (minor) |
| `src/pages/tidrapport.astro` | Major rewrite |
