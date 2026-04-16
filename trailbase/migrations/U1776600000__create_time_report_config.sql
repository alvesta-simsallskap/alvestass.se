-- GDPR: No personal data — operational configuration only
-- Access: Authenticated read (service user token); admin-only write

CREATE TABLE IF NOT EXISTS time_report_config (
  id                    INTEGER PRIMARY KEY,
  active_month_key      TEXT    NOT NULL CHECK(active_month_key != ''),
  active_month_display  TEXT    NOT NULL CHECK(active_month_display != ''),
  extra_time_simskola   INTEGER NOT NULL DEFAULT 30,
  extra_time_training   INTEGER NOT NULL DEFAULT 15,
  half_day_salary       INTEGER NOT NULL DEFAULT 500,
  full_day_salary       INTEGER NOT NULL DEFAULT 1000,
  overnight_salary      INTEGER NOT NULL DEFAULT 300
) STRICT;

INSERT INTO time_report_config (
  id, active_month_key, active_month_display
) VALUES (
  1, '2026-04', 'april 2026'
);
