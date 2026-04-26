# Data Model: Two-Step Time Report Wizard

**Feature**: 005-time-report-wizard  
**Date**: 2026-04-26

---

## Entity: Instructor (updated)

**Storage**: Trailbase table `instructors` (SQLite STRICT)  
**Cardinality**: Many rows (~30 active instructors)  
**Access**: Authenticated read (service user token); admin-only write  
**GDPR**: Personal data — see notes below

### Changes from 002-time-report-trailbase

| Column | Before | After | Reason |
|--------|--------|-------|--------|
| `swim_school_rate` | `INTEGER NOT NULL` | `INTEGER` (nullable) | Support coach-only instructors (FR-013) |
| `travel_compensation` | (not present) | `INTEGER NOT NULL DEFAULT 0` | Eligibility flag for Milersättning (FR-012) |

### GDPR Documentation

| Field | Personal data? | Legal basis | Retention |
|-------|---------------|-------------|-----------|
| `email` | Yes — contact identifier | Contractual necessity (employment) | Until end of employment + 1 year |
| `swim_school_rate` | Yes — salary data | Contractual necessity (employment) | Until end of employment + 1 year |
| `coach_rate` | Yes — salary data | Contractual necessity (employment) | Until end of employment + 1 year |
| `travel_compensation` | Yes — employment term | Contractual necessity (employment) | Until end of employment + 1 year |

Must be registered in `docs/gdpr-register.md` (TODO: GDPR_REGISTER) before
go-live. Deletion path: admin deletes row via Trailbase admin UI.

### Full Field Listing (post-migration)

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `email` | TEXT | NOT NULL UNIQUE | Lowercase email address (lookup key) |
| `swim_school_rate` | INTEGER | DEFAULT NULL | SEK/h for swim school; NULL if no swim school duties |
| `coach_rate` | INTEGER | DEFAULT NULL | SEK/h for coaching; NULL if no coaching duties |
| `travel_compensation` | INTEGER | NOT NULL DEFAULT 0 | 1 = eligible for Milersättning; 0 = not eligible |

**Invariant**: `swim_school_rate IS NOT NULL OR coach_rate IS NOT NULL` —
at least one rate must be set (enforced by `CHECK` constraint).

### SQL Migration

```sql
-- trailbase/migrations/U1776686400__update_instructors.sql
-- Changes: make swim_school_rate nullable; add travel_compensation boolean
-- GDPR: travel_compensation is personal data (employment term)
-- Legal basis: Contractual necessity (employment relationship)
-- Retention: Until end of employment + 1 year
-- Deletion path: Admin deletes row via Trailbase admin UI

-- SQLite cannot ALTER COLUMN — recreate table with new schema
CREATE TABLE instructors_new (
  id                  INTEGER PRIMARY KEY,
  email               TEXT    NOT NULL UNIQUE CHECK(email != ''),
  swim_school_rate    INTEGER CHECK(swim_school_rate IS NULL OR swim_school_rate > 0),
  coach_rate          INTEGER CHECK(coach_rate IS NULL OR coach_rate > 0),
  travel_compensation INTEGER NOT NULL DEFAULT 0,
  CHECK(swim_school_rate IS NOT NULL OR coach_rate IS NOT NULL)
) STRICT;

INSERT INTO instructors_new (id, email, swim_school_rate, coach_rate, travel_compensation)
SELECT id, email, swim_school_rate, coach_rate, 0
FROM instructors;

DROP TABLE instructors;
ALTER TABLE instructors_new RENAME TO instructors;
```

### TypeScript Interface (updated)

```typescript
// src/lib/types.ts

export interface Instructor {
  id: number;
  email: string;
  swim_school_rate: number | null;  // null for coach-only instructors
  coach_rate: number | null;
  travel_compensation: boolean;      // true = Milersättning eligible
}
```

---

## Derived Type: InstructorRole

Not stored — computed at runtime from the `Instructor` record by
`/api/lookup-instructor`. Only this shape is sent to the browser.

```typescript
// src/pages/api/lookup-instructor.ts (internal type)

interface InstructorRole {
  swimSchool: boolean;          // swim_school_rate IS NOT NULL
  coach: boolean;               // coach_rate IS NOT NULL
  travelCompensation: boolean;  // travel_compensation === 1
}
```

---

## No new tables

This feature adds no new Trailbase tables. All changes are to the existing
`instructors` table.

---

## Data Flow Summary

```
Browser (step 1)
  └── POST /api/lookup-instructor { email }
        └── Cloudflare Worker
              └── Trailbase authenticateServiceUser → fetchInstructor(email)
                    └── returns InstructorRole { swimSchool, coach, travelCompensation }
                          → stored in Alpine.js state as `role`
                          → step 2 form sections rendered conditionally

Browser (step 2 submit)
  └── POST /api/send-time-report (existing flow, unchanged)
        └── Cloudflare Worker
              └── Trailbase fetchInstructor(email) for salary calculation
                    └── email is embedded as hidden field from Alpine state
```
