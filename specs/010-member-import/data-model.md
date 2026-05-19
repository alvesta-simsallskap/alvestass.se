# Data Model: Member Register

**Branch**: `010-member-import` | **Date**: 2026-05-19  
**Migration file**: `trailbase/migrations/U1779235200__create_member_register.sql`

## Entity Relationship Overview

```
members (iid PK)
  ├── guardians.member_iid  (N guardians per swimmer)
  ├── member_training_groups.member_iid  (M:N → training_groups)
  └── family_members.member_iid  (M:N → families)

training_groups (id PK)
  └── member_training_groups.group_id

families (id PK)
  └── family_members.family_id
```

## Tables

### `members`

Stores every active person in the club: swimmers, instructors, and board members.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `iid` | TEXT | PRIMARY KEY | IdrottsID, e.g. `IID12345678` |
| `first_name` | TEXT | NOT NULL | |
| `last_name` | TEXT | NOT NULL | |
| `gender` | TEXT | nullable | `'Man'` \| `'Kvinna'` |
| `date_of_birth` | TEXT | nullable | `YYYY-MM-DD` |
| `city` | TEXT | nullable | Place of residence |
| `member_since` | TEXT | nullable | `YYYY-MM-DD` |
| `email` | TEXT | nullable | Primary contact email; used for future login |
| `phone` | TEXT | nullable | Mobile phone number |
| `is_swimmer` | INTEGER | NOT NULL DEFAULT 0 | Boolean (0/1) |
| `is_instructor` | INTEGER | NOT NULL DEFAULT 0 | Boolean (0/1) |
| `is_board_member` | INTEGER | NOT NULL DEFAULT 0 | Boolean (0/1) |

**GDPR annotation**:
- Personal data: name, date of birth, city, email, phone
- Legal basis: Art. 6(1)(b) contractual necessity (membership agreement)
- Retention: Until membership ends + 2 years; review annually
- Deletion path: `alvestass-admin delete-member --iid <IID>` (cascades to guardians, member_training_groups, family_members)

---

### `guardians`

Legal guardians of active minor swimmers. One swimmer may have up to three guardians.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `member_iid` | TEXT | NOT NULL, FK → members | The swimmer this guardian is responsible for |
| `member_iid_self` | TEXT | nullable, FK → members | Set if the guardian is also a club member |
| `first_name` | TEXT | NOT NULL | |
| `last_name` | TEXT | NOT NULL | |
| `phone` | TEXT | nullable | |
| `email` | TEXT | nullable | |

**GDPR annotation**:
- Personal data: name, phone, email
- Legal basis: Art. 6(1)(f) legitimate interest (safety; minor member management)
- Retention: As long as linked swimmer is an active member + 2 years
- Deletion path: CASCADE on `members.iid` deletion; or direct row delete via Trailbase admin UI

---

### `training_groups`

Named swim groups. Time slots are deliberately excluded — they belong in a future schedule feature.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `name` | TEXT | NOT NULL UNIQUE | e.g. `Baddaren`, `A-gruppen` |
| `category` | TEXT | NOT NULL | `swim_school` \| `adult` \| `masters` \| `competitive` \| `technique` |

No personal data. Retained indefinitely (operational data).

---

### `member_training_groups`

Join table connecting members to their training groups with a role.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `member_iid` | TEXT | NOT NULL, FK → members | |
| `group_id` | INTEGER | NOT NULL, FK → training_groups | |
| `role` | TEXT | NOT NULL | `participant` \| `instructor` \| `head_instructor` |

Unique constraint: `(member_iid, group_id)`.

No personal data beyond the foreign key reference. Retained as long as the member record exists.

---

### `families`

Represents a family unit. Used to identify sibling relationships and co-locate guardian communications.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `source_label` | TEXT | nullable | The `Familj` string from IdrottOnline (e.g. `"22011829"`) for traceability |

No personal data in this table itself.

---

### `family_members`

Join table connecting members to families.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `family_id` | INTEGER | NOT NULL, FK → families | |
| `member_iid` | TEXT | NOT NULL, FK → members | |

Unique constraint: `(family_id, member_iid)`.

---

## Full SQL Migration

```sql
-- Member register: six tables for active members, guardians, training groups, and families.
-- GDPR: See individual table comments for legal basis and retention details.
-- Production import is GATED on:
--   (1) /integritetspolicy page published at alvestass.se/integritetspolicy
--   (2) docs/gdpr-register.md updated with entries for all six tables below

-- members: Active swimmers, instructors, and board members.
-- GDPR personal data: first_name, last_name, gender, date_of_birth, city, email, phone
-- Legal basis: Art. 6(1)(b) contractual necessity (membership agreement)
-- Retention: Until membership ends + 2 years, then delete or anonymise
-- Access: Authenticated only — no public API access
-- Deletion path: alvestass-admin delete-member --iid <IID> (cascades to child tables)
CREATE TABLE IF NOT EXISTS members (
  iid             TEXT    PRIMARY KEY,
  first_name      TEXT    NOT NULL CHECK(first_name != ''),
  last_name       TEXT    NOT NULL CHECK(last_name != ''),
  gender          TEXT    CHECK(gender IN ('Man', 'Kvinna')),
  date_of_birth   TEXT,
  city            TEXT,
  member_since    TEXT,
  email           TEXT,
  phone           TEXT,
  is_swimmer      INTEGER NOT NULL DEFAULT 0 CHECK(is_swimmer IN (0, 1)),
  is_instructor   INTEGER NOT NULL DEFAULT 0 CHECK(is_instructor IN (0, 1)),
  is_board_member INTEGER NOT NULL DEFAULT 0 CHECK(is_board_member IN (0, 1))
) STRICT;

-- guardians: Legal guardians of active minor swimmers.
-- GDPR personal data: first_name, last_name, phone, email
-- Legal basis: Art. 6(1)(f) legitimate interest (minor member safety and communication)
-- Retention: As long as linked swimmer is active + 2 years
-- Access: Authenticated only
-- Deletion path: CASCADE on members.iid deletion; or direct row delete via Trailbase admin UI
CREATE TABLE IF NOT EXISTS guardians (
  id              INTEGER PRIMARY KEY,
  member_iid      TEXT    NOT NULL REFERENCES members(iid) ON DELETE CASCADE,
  member_iid_self TEXT    REFERENCES members(iid),
  first_name      TEXT    NOT NULL CHECK(first_name != ''),
  last_name       TEXT    NOT NULL CHECK(last_name != ''),
  phone           TEXT,
  email           TEXT
) STRICT;

-- training_groups: Named swim groups without time slots.
-- No personal data. Retained indefinitely.
CREATE TABLE IF NOT EXISTS training_groups (
  id       INTEGER PRIMARY KEY,
  name     TEXT    NOT NULL UNIQUE CHECK(name != ''),
  category TEXT    NOT NULL CHECK(category IN ('swim_school', 'adult', 'masters', 'competitive', 'technique'))
) STRICT;

-- member_training_groups: Member ↔ training group join table with role.
-- No personal data beyond FK reference. Retained as long as member record exists.
CREATE TABLE IF NOT EXISTS member_training_groups (
  id         INTEGER PRIMARY KEY,
  member_iid TEXT    NOT NULL REFERENCES members(iid) ON DELETE CASCADE,
  group_id   INTEGER NOT NULL REFERENCES training_groups(id) ON DELETE CASCADE,
  role       TEXT    NOT NULL CHECK(role IN ('participant', 'instructor', 'head_instructor')),
  UNIQUE(member_iid, group_id)
) STRICT;

-- families: Family unit records for grouping sibling members.
-- No personal data.
CREATE TABLE IF NOT EXISTS families (
  id           INTEGER PRIMARY KEY,
  source_label TEXT
) STRICT;

-- family_members: Member ↔ family join table.
-- No personal data beyond FK reference.
CREATE TABLE IF NOT EXISTS family_members (
  id         INTEGER PRIMARY KEY,
  family_id  INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  member_iid TEXT    NOT NULL REFERENCES members(iid) ON DELETE CASCADE,
  UNIQUE(family_id, member_iid)
) STRICT;
```

## Validation Rules

| Rule | Table | Constraint |
|------|-------|------------|
| IID must be non-empty | members | PRIMARY KEY NOT NULL |
| At least one role must be set | members | Enforced by import logic (not DB constraint) |
| Category must be a known value | training_groups | CHECK constraint |
| Role must be a known value | member_training_groups | CHECK constraint |
| Member-group pair is unique | member_training_groups | UNIQUE constraint |
| Member-family pair is unique | family_members | UNIQUE constraint |
| Guardian must reference an existing member | guardians | FOREIGN KEY + ON DELETE CASCADE |

## Source Field Mapping

### IdrottOnline (`ExportedPersons-3.csv`) → `members`

| Source column | Target field | Notes |
|---------------|-------------|-------|
| `IdrottsID` | `iid` | e.g. `IID12345678` |
| `Förnamn` | `first_name` | |
| `Efternamn` | `last_name` | |
| `Kön` | `gender` | `'Man'` or `'Kvinna'` |
| `Födelsedat./Personnr.` | `date_of_birth` | Take first 10 chars: `YYYY-MM-DD` (note: source format is `YYYYMMDD-xxxx`) |
| `Kontaktadress - Postort` | `city` | |
| `Medlem sedan` | `member_since` | |
| `E-post kontakt` | `email` | |
| `Telefon mobil` | `phone` | |
| `Roller` | `is_board_member` | Set 1 if contains `Styrelseledamot`, `Ordförande`, `Kassör`, `Sekreterare`, `Vice ordförande` |

### WeUnite (`Grupplista`) → `member_training_groups`

| Source column | Target field | Notes |
|---------------|-------------|-------|
| `Personnummer` | (join key — not stored) | Used to look up IID from IdrottOnline |
| `Grupp` (normalised) | `training_groups.name` | Time slot stripped by regex |
| `Roll` | `role` | `Deltagare` → `participant`; `Ledare` → `instructor`; `Huvudledare` → `head_instructor` |
| `Sektion` / `Nivå` | `training_groups.category` | Mapped per research.md §4 |

### WeUnite → `guardians`

| Source columns | Target field | Notes |
|----------------|-------------|-------|
| `Målsman N, Förnamn` | `first_name` | N = 1, 2, 3 |
| `Målsman N, Efternamn` | `last_name` | |
| `Målsman N, Telefon` | `phone` | |
| `Målsman N, E-post` | `email` | |
| Swimmer's IID (from join) | `member_iid` | |

### IdrottOnline `Familj` field → `families` + `family_members`

Members sharing the same non-empty `Familj` value are grouped into a single `families` row. The raw `Familj` string is stored as `source_label` for traceability.
