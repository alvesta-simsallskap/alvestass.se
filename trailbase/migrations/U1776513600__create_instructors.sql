-- GDPR: Personal data — email and salary rates
-- Legal basis: Contractual necessity (employment relationship)
-- Retention: Until end of employment + 1 year for accounting
-- Access: Authenticated read (service user token); admin-only write
-- Deletion path: Admin deletes row via Trailbase admin UI

CREATE TABLE IF NOT EXISTS instructors (
  id               INTEGER PRIMARY KEY,
  email            TEXT    NOT NULL UNIQUE CHECK(email != ''),
  swim_school_rate INTEGER NOT NULL CHECK(swim_school_rate > 0),
  coach_rate       INTEGER CHECK(coach_rate IS NULL OR coach_rate > 0)
) STRICT;
