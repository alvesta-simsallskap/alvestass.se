-- GDPR: No personal data — training schedule only
-- Access: Authenticated read (service user token); admin-only write

CREATE TABLE IF NOT EXISTS time_report_sessions (
  id             INTEGER PRIMARY KEY,
  month_key      TEXT    NOT NULL CHECK(month_key != ''),
  training_group TEXT    NOT NULL CHECK(training_group IN (
                   'simskola', 'tavlingA', 'tavlingB',
                   'teknik', 'masters', 'vuxencrawl'
                 )),
  date           TEXT    NOT NULL CHECK(date != ''),
  title          TEXT    NOT NULL CHECK(title != ''),
  hours          INTEGER NOT NULL CHECK(hours >= 0),
  minutes        INTEGER NOT NULL DEFAULT 0 CHECK(minutes >= 0 AND minutes < 60)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_sessions_month_group
  ON time_report_sessions (month_key, training_group);
