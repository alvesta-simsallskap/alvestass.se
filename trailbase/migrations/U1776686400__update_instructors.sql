-- trailbase/migrations/U1776686400__update_instructors.sql
-- Changes: make swim_school_rate nullable (supports coach-only instructors);
--          add travel_compensation boolean (Milersättning eligibility).
--
-- GDPR: travel_compensation is personal data (employment term — whether employer
--       covers travel costs is a contractual obligation).
-- Legal basis: Contractual necessity (Art. 6(1)(b) GDPR — employment relationship)
-- Retention: Until end of employment + 1 year for accounting purposes
-- Access: Authenticated read (service user); admin-only write
-- Deletion path: Admin deletes instructor row via Trailbase admin UI
-- TODO(GDPR_REGISTER): travel_compensation must be added to docs/gdpr-register.md
--                      before go-live (spec 005-time-report-wizard)
--
-- SQLite does not support ALTER COLUMN — use table-rename recreation pattern.
-- This migration runs inside a transaction; a failure mid-way rolls back cleanly.

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
