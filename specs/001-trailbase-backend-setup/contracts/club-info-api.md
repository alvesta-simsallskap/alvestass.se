# API Contract: ClubInfo — Trailbase Records Endpoint

**Feature**: 001-trailbase-backend-setup
**Date**: 2026-04-16
**Base URL**: `https://<app-name>.fly.dev` (set as `TRAILBASE_URL` Cloudflare secret)

> **Note**: Trailbase auto-generates REST endpoints from table definitions. The exact path prefix (`/api/collections/v1/` or `/api/records/v1/`) MUST be verified from the Trailbase admin UI → API docs tab after first deployment. The contract below uses the expected path; adjust if the actual path differs.

---

## GET /api/collections/v1/club_info

Returns the list of `club_info` records. Since there is exactly one record, this always returns a single-element list.

**Authentication**: None (table configured with public read access)

**Query parameters**:

| Param | Value | Notes |
|-------|-------|-------|
| `limit` | `1` | Retrieve only the first (and only) record |

**Example request** (from Cloudflare Worker):
```
GET https://<app-name>.fly.dev/api/collections/v1/club_info?limit=1
```

**Success response** — `200 OK`:
```json
{
  "records": [
    {
      "id": 1,
      "name": "Alvesta Simsällskap",
      "tagline": "Simning för alla",
      "founding_year": 1921,
      "short_description": "Alvesta Simsällskap grundades 1921...",
      "address": "Ekskogsvägen 4",
      "city": "Alvesta",
      "postal_code": "342 30",
      "phone": "076 027 94 10",
      "email": "kansli@alvestass.se"
    }
  ],
  "cursor": null
}
```

> **Important**: The exact top-level key name (`records`, `data`, or similar) MUST be confirmed from the live Trailbase API response during implementation. Update `src/lib/trailbase.ts` accordingly.

**Empty database** — `200 OK`:
```json
{ "records": [], "cursor": null }
```
The Astro page MUST handle an empty array gracefully (show placeholder text, not crash).

**Backend unreachable** — network error / timeout:
The Cloudflare Worker fetch will throw. The page handler MUST catch this and return a response with `Cache-Control: public, stale-if-error=86400` so Cloudflare serves the last cached version. See quickstart.md for the error-handling pattern.

---

## GET /api/collections/v1/club_info/1

Fetch the single record directly by ID. Alternative to the list endpoint.

**Authentication**: None

**Success response** — `200 OK`:
```json
{
  "id": 1,
  "name": "Alvesta Simsällskap",
  ...
}
```

**Not found** — `404 Not Found`:
Handle identically to the empty-array case above.

---

## Trailbase Admin Endpoints (not called by the website)

The Trailbase admin UI at `https://<app-name>.fly.dev/_/` handles all write operations (create, update records). These are not part of the website's integration surface and require admin credentials.
