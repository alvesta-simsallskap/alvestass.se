# Data Model: Trailbase Backend Setup (Minimal Starter)

**Feature**: 001-trailbase-backend-setup
**Date**: 2026-04-16

---

## Entity: ClubInfo

**Storage**: Trailbase table `club_info` (SQLite)
**Cardinality**: Exactly one row (single-record pattern)
**Access**: Public read (no auth); admin-only write

### Fields

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Trailbase auto-assigns; the single row will have id `1` |
| `name` | TEXT | NOT NULL | Club's official name, e.g. "Alvesta Simsällskap" |
| `tagline` | TEXT | NOT NULL, DEFAULT '' | Short motto or descriptor |
| `founding_year` | INTEGER | NOT NULL | e.g. 1921 |
| `short_description` | TEXT | NOT NULL, DEFAULT '' | ≤ 300 characters; shown as intro text |
| `address` | TEXT | NOT NULL | Street address (street + number) |
| `city` | TEXT | NOT NULL | e.g. "Alvesta" |
| `postal_code` | TEXT | NOT NULL | Swedish format: "342 30" |
| `phone` | TEXT | NOT NULL | Public phone in display format: "076 027 94 10" |
| `email` | TEXT | NOT NULL | Public contact email, e.g. "kansli@alvestass.se" |

### Validation Rules (enforced in Trailbase admin or migration constraints)

- `name`, `address`, `city`, `postal_code`, `phone`, `email` MUST NOT be empty (NOT NULL + CHECK constraint or Trailbase required-field rule)
- `founding_year` MUST be a plausible year (CHECK: founding_year >= 1800 AND founding_year <= 2100)
- `short_description` MUST be ≤ 300 characters (CHECK: length(short_description) <= 300)

### SQL Migration (canonical source of truth)

```sql
-- trailbase/migrations/0001_initial.sql
CREATE TABLE club_info (
  id           INTEGER PRIMARY KEY,
  name         TEXT    NOT NULL CHECK(name         != ''),
  tagline      TEXT    NOT NULL DEFAULT '',
  founding_year INTEGER NOT NULL CHECK(founding_year >= 1800 AND founding_year <= 2100),
  short_description TEXT NOT NULL DEFAULT '' CHECK(length(short_description) <= 300),
  address      TEXT    NOT NULL CHECK(address      != ''),
  city         TEXT    NOT NULL CHECK(city         != ''),
  postal_code  TEXT    NOT NULL CHECK(postal_code  != ''),
  phone        TEXT    NOT NULL CHECK(phone        != ''),
  email        TEXT    NOT NULL CHECK(email        != '')
);

-- Seed row: Alvesta Simsällskap contact details
INSERT INTO club_info (
  id, name, tagline, founding_year, short_description,
  address, city, postal_code, phone, email
) VALUES (
  1,
  'Alvesta Simsällskap',
  '',
  1921,
  '',
  '',       -- fill in before deploy
  'Alvesta',
  '342 30',
  '076 027 94 10',
  'kansli@alvestass.se'
);
```

> **Note**: The `address`, `tagline`, and `short_description` seed values are intentionally left blank here — the admin MUST fill them in via the Trailbase admin UI before going live (see quickstart.md). Storing the physical address in git is acceptable (it is public info), but fill it in during deployment rather than hardcoding it here.

---

## GDPR Documentation

| Field | Personal data? | Legal basis | Retention |
|-------|---------------|-------------|-----------|
| All fields | No — organizational contact info only | Legitimate interest (public organizational identity) | Indefinite |

No GDPR data-subject rights procedures are required for this entity.

---

## TypeScript Interface (for `src/lib/trailbase.ts`)

```typescript
export interface ClubInfo {
  id: number;
  name: string;
  tagline: string;
  founding_year: number;
  short_description: string;
  address: string;
  city: string;
  postal_code: string;
  phone: string;
  email: string;
}
```
