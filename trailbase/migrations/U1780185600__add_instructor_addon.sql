-- trailbase/migrations/U1780185600__add_instructor_addon.sql
-- Changes: add a generic per-instructor monetary addon (addon_amount + addon_description),
--          e.g. a fixed travel compensation ("Reseersättning", 895 kr) automatically added
--          to the time-report salary total. Both columns are set or both NULL.
--
-- GDPR: addon_amount/addon_description are personal data (employment term — a salary
--       supplement the employer pays is a contractual obligation).
-- Legal basis: Contractual necessity (Art. 6(1)(b) GDPR — employment relationship)
-- Retention: Until end of employment + 1 year for accounting purposes
-- Access: Authenticated read (service user); admin-only write
-- Deletion path: Admin deletes instructor row via Trailbase admin UI
-- TODO(GDPR_REGISTER): addon_amount/addon_description added to docs/gdpr-register.md
--
-- SQLite cannot add a cross-column CHECK via ALTER, so use the table-rename recreation
-- pattern (as in U1776686400). This migration runs inside a transaction; a failure
-- mid-way rolls back cleanly.

CREATE TABLE instructors_new (
  id                  INTEGER PRIMARY KEY,
  email               TEXT    NOT NULL UNIQUE CHECK(email != ''),
  name                TEXT    NOT NULL DEFAULT '',
  swim_school_rate    INTEGER CHECK(swim_school_rate IS NULL OR swim_school_rate > 0),
  coach_rate          INTEGER CHECK(coach_rate IS NULL OR coach_rate > 0),
  travel_compensation INTEGER NOT NULL DEFAULT 0,
  addon_amount        INTEGER CHECK(addon_amount IS NULL OR addon_amount > 0),
  addon_description   TEXT,
  CHECK(swim_school_rate IS NOT NULL OR coach_rate IS NOT NULL),
  CHECK(
    (addon_amount IS NULL AND addon_description IS NULL)
    OR (addon_amount IS NOT NULL AND addon_description IS NOT NULL AND addon_description != '')
  )
) STRICT;

INSERT INTO instructors_new (id, email, name, swim_school_rate, coach_rate, travel_compensation)
SELECT id, email, name, swim_school_rate, coach_rate, travel_compensation
FROM instructors;

DROP TABLE instructors;
ALTER TABLE instructors_new RENAME TO instructors;
