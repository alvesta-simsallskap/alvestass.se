# Quickstart: Time Report Trailbase Migration

**Feature**: 002-time-report-trailbase
**Date**: 2026-04-16

---

## Prerequisites

- Trailbase deployed and accessible at `https://alvestass-trailbase.fly.dev`
- Admin credentials for the Trailbase admin UI
- Cloudflare Worker secrets configured (`wrangler secret put`)

---

## 1. Apply the Three New Migrations

The three migration files are baked into the Docker image. They will be applied
automatically on the next `fly deploy`:

```bash
cd trailbase/
fly deploy
```

After deploy, verify in the Trailbase admin UI (`https://alvestass-trailbase.fly.dev/_/admin/`):
- `time_report_config` table exists with 1 seed row
- `time_report_sessions` table exists (empty)
- `instructors` table exists (empty)

---

## 2. Configure Table Access Rules

In the Trailbase admin UI, set access rules for each table:

| Table | Read | Write/Update/Delete |
|-------|------|---------------------|
| `time_report_config` | Authenticated | Admin only |
| `time_report_sessions` | Authenticated | Admin only |
| `instructors` | Authenticated | Admin only |

---

## 3. Create the Service User

The Cloudflare Worker needs a dedicated Trailbase user to authenticate for all Trailbase API calls.

In the Trailbase admin UI → Users → Create user:
- Email: `service@alvestass.se` (or any internal email)
- Password: generate a strong random password
- Role: regular user (not admin)

This user gets "Authenticated" access to the `instructors`, `time_report_config`,
and `time_report_sessions` tables (all require authentication per FR-008).

---

## 4. Set Cloudflare Worker Secrets

```bash
wrangler secret put TRAILBASE_SERVICE_EMAIL
# Enter: service@alvestass.se

wrangler secret put TRAILBASE_SERVICE_PASSWORD
# Enter: <the password from step 3>
```

Verify:
```bash
wrangler secret list
```

---

## 5. Enter the Current Month's Schedule

In the Trailbase admin UI → `time_report_sessions` table, add rows for the
current reporting month. Each row represents one session:

| Field | Value |
|-------|-------|
| `month_key` | `2026-04` (must match `time_report_config.active_month_key`) |
| `group` | `simskola`, `tavlingA`, `tavlingB`, `teknik`, `masters`, or `vuxencrawl` |
| `date` | ISO date, e.g. `2026-04-07` |
| `title` | e.g. `Träning`, `Simskola`, `ÖGP Växjö` |
| `hours` | Duration hours (10=half-day, 15=overnight, 20=full-day, else normal) |
| `minutes` | Duration minutes (0–59) |

The sessions from `time-report-items.json` for the current month must be
entered here. Historical months do not need to be migrated.

---

## 6. Enter Instructor Data

In the Trailbase admin UI → `instructors` table, add one row per instructor:

| Field | Value |
|-------|-------|
| `email` | Instructor's email address (lowercase) |
| `swim_school_rate` | Hourly rate in SEK for swim school sessions |
| `coach_rate` | Hourly rate in SEK for coaching; leave blank if no coaching duties |

The 26 instructors currently in `send-time-report.ts` must be transferred here.

> **GDPR note**: These are personal records. Only enter instructors currently
> under active agreements. Do not enter historical instructors unless required
> for accounting. The `/integritetspolicy` page MUST be live before this step.

---

## 7. Verify the API

```bash
# Config (should return the seed row)
curl "https://alvestass-trailbase.fly.dev/api/records/v1/time_report_config?limit=1"

# Sessions (should return rows you entered in step 5)
curl "https://alvestass-trailbase.fly.dev/api/records/v1/time_report_sessions?filter[month_key][\$eq]=2026-04&limit=10"

# Instructor lookup — requires auth (test manually via admin UI or Trailbase API)
```

---

## 8. Test the Form Locally

```bash
pnpm dev
open http://localhost:4321/tidrapport
```

Expected:
- Form heading shows correct month name from Trailbase config
- Session checkboxes show sessions entered in step 5
- Submit form with your own email → preliminary salary estimate in HTML preview

---

## 9. Build Gate

```bash
pnpm build
```

Expected: zero TypeScript errors, zero `astro check` warnings.

---

## Changing the Active Month (Monthly Admin Task)

Each month:
1. Add new sessions to `time_report_sessions` for the new month key
2. Update `time_report_config.active_month_key` and `active_month_display`
3. The form automatically reflects the new month — no deployment required

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| Form shows "Tidrapport laddas inte just nu" | Trailbase unreachable or no config row | Check fly.io status; verify seed row in `time_report_config` |
| Sessions list is empty | No rows in `time_report_sessions` for active month | Enter sessions in admin UI; verify `month_key` matches config |
| Salary estimate missing from email | Service auth failed or email not in `instructors` | Check `TRAILBASE_SERVICE_EMAIL/PASSWORD` secrets; verify instructor row exists |
| Build fails with TS errors | Types mismatch after refactor | Run `pnpm build` and fix reported errors |
