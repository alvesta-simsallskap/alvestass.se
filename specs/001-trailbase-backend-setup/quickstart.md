# Quickstart: Trailbase Backend Setup (Minimal Starter)

**Feature**: 001-trailbase-backend-setup
**Date**: 2026-04-16

---

## Prerequisites

- `flyctl` installed and authenticated (`fly auth login`)
- Docker installed (for local Trailbase testing)
- Cloudflare account with the `alvestass-se` Worker project

---

## 1. Deploy Trailbase to fly.io

```bash
# From the repository root
cd trailbase/

# Create the fly.io app (first time only)
fly launch --no-deploy --region arn --name alvestass-trailbase

# Create the persistent volume for SQLite (first time only)
fly volumes create trailbase_data --region arn --size 1

# Deploy
fly deploy
```

---

## 2. Retrieve Admin Credentials

On first startup Trailbase automatically creates an admin user and prints the
credentials to stdout. Retrieve them from the fly.io logs:

```bash
fly logs --app alvestass-trailbase
```

Look for output similar to:

```
Created admin user: admin@localhost  password: <generated-password>
```

Save the password — you'll need it to log in to the admin UI at:
```
https://alvestass-trailbase.fly.dev/_/admin/
```

To change the password later:
```bash
fly ssh console --app alvestass-trailbase
trail user reset-password admin@localhost <new-password>
```

---

## 3. First-time Trailbase Setup

1. Open `https://alvestass-trailbase.fly.dev/_/admin/` and log in with the
   credentials from step 2.
2. Navigate to **Tables** → the `club_info` table should already exist (created
   by the migration seeded from the Docker image).
3. Verify the seed row is present and fill in the blank fields:
   - `address`: the pool's street address
   - `tagline`: a short motto (optional but recommended)
   - `short_description`: 1–2 sentences about the club (≤ 300 chars)
4. In the `club_info` table view, click **"Expose an API for this table"** and configure:
   - **Read**: `Everyone` (public, unauthenticated)
   - **Write/Update/Delete**: `Admin only`

---

## 4. Confirm the REST API

```bash
curl "https://alvestass-trailbase.fly.dev/api/records/v1/club_info?limit=1"
```

Expected: a JSON object with a `records` array containing the club_info row.

> **Note**: The correct path prefix is `/api/records/v1/` — not `/api/collections/v1/`.
> `src/lib/trailbase.ts` already uses the correct path.

---

## 5. Configure the Cloudflare Worker Secret

```bash
# From the repository root
wrangler secret put TRAILBASE_URL
# When prompted, enter: https://alvestass-trailbase.fly.dev
```

Verify:
```bash
wrangler secret list
```

---

## 6. Verify the Website Integration

```bash
# Start local dev server (requires .dev.vars with TRAILBASE_URL set)
pnpm dev

open http://localhost:4321/kontakt
```

Expected: club name, address, phone, email, and short description from Trailbase.

**Test the 5-minute update flow**:
1. Change a field (e.g. the tagline) in the Trailbase admin UI
2. Wait up to 5 minutes for the Cloudflare edge cache to expire
3. Hard-refresh `/kontakt` — confirm the new value appears

**Test the fallback**:
1. `fly scale count 0 --app alvestass-trailbase`
2. Visit `/kontakt` — it should serve the last cached version (not error)
3. Restore: `fly scale count 1 --app alvestass-trailbase`

---

## 7. Build Gate

```bash
pnpm build
```

Expected: zero TypeScript errors, zero `astro check` errors, successful build.

---

## Starting Fresh (Reset)

If the Trailbase state is corrupted or you need to start from scratch:

```bash
# 1. Stop the machine
fly scale count 0 --app alvestass-trailbase

# 2. Find the volume ID
fly volumes list --app alvestass-trailbase

# 3. Destroy the volume
fly volumes destroy <vol-id>

# 4. Recreate the volume
fly volumes create trailbase_data --region arn --size 1

# 5. Redeploy — migrations are re-seeded automatically by the entrypoint
fly deploy

# 6. Retrieve new admin credentials from logs
fly logs --app alvestass-trailbase
```

---

## Key Patterns Reference

### Fetching ClubInfo in an SSR Astro page

```typescript
// src/pages/kontakt.astro (frontmatter)
import { fetchClubInfo } from '../lib/trailbase';

export const prerender = false;

let clubInfo = null;
try {
  clubInfo = await fetchClubInfo(Astro.locals.runtime.env.TRAILBASE_URL);
} catch {
  // Trailbase unreachable — Cloudflare stale-if-error serves last cached page
}

Astro.response.headers.set(
  'Cache-Control',
  'public, max-age=300, stale-if-error=86400'
);
```

### env.d.ts addition

```typescript
TRAILBASE_URL: string;   // https://alvestass-trailbase.fly.dev
```

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| No admin credentials in logs | Machine restarted but not first run | Reset the volume (see Starting Fresh) |
| curl returns 404 on `/api/records/v1/club_info` | API not exposed or read access not set to Everyone | Click "Expose an API for this table" in admin UI and set read to Everyone |
| `club_info` table missing after deploy | Migration seeding failed | Check `fly logs` for entrypoint errors; verify migration filename matches `U<timestamp>__<name>.sql` |
| Page shows blank data in dev | `TRAILBASE_URL` not set locally | Add to `.dev.vars`: `TRAILBASE_URL=https://alvestass-trailbase.fly.dev` |
| fly deploy fails | No volume attached | Run `fly volumes create` step again |
| Data not updating after 5 min | Cloudflare cache not purging | Check `Cache-Control` header is present; purge manually from Cloudflare dashboard |
