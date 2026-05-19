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
  id              INTEGER PRIMARY KEY,
  iid             TEXT    NOT NULL UNIQUE CHECK(iid != ''),
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

-- member_training_groups: Member <-> training group join table with role.
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

-- family_members: Member <-> family join table.
-- No personal data beyond FK reference.
CREATE TABLE IF NOT EXISTS family_members (
  id         INTEGER PRIMARY KEY,
  family_id  INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  member_iid TEXT    NOT NULL REFERENCES members(iid) ON DELETE CASCADE,
  UNIQUE(family_id, member_iid)
) STRICT;
