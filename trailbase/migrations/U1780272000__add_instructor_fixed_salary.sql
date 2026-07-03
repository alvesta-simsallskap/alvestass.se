-- trailbase/migrations/U1780272000__add_instructor_fixed_salary.sql
-- Changes: add fixed-salary support to instructors. fixed_salary marks instructors paid a
--          fixed monthly wage (no hourly estimate); time_bank is their personal balance in
--          minutes (signed), adjusted manually by the reviewer at review time. For these
--          instructors the time report shows "Fast lön" + the estimated time-bank change
--          instead of a salary figure.
--
-- GDPR: fixed_salary/time_bank are personal data (employment terms — salary model and
--       accrued time balance are contractual matters).
-- Legal basis: Contractual necessity (Art. 6(1)(b) GDPR — employment relationship)
-- Retention: Until end of employment + 1 year for accounting purposes
-- Access: Authenticated read (service user); admin-only write
-- Deletion path: Admin deletes instructor row via Trailbase admin UI
-- TODO(GDPR_REGISTER): fixed_salary/time_bank added to docs/gdpr-register.md
--
-- Both columns are single-column additions with defaults and no cross-column CHECK, so a plain
-- ALTER TABLE ADD COLUMN is sufficient (no table-rename recreation needed; cf. U1780185600 which
-- required it only for a cross-column CHECK). SQLite allows a single-column CHECK on ADD COLUMN.

ALTER TABLE instructors ADD COLUMN fixed_salary INTEGER NOT NULL DEFAULT 0 CHECK(fixed_salary IN (0, 1));
ALTER TABLE instructors ADD COLUMN time_bank    INTEGER NOT NULL DEFAULT 0;  -- signed balance in minutes
