-- Initial schema: club_info table for public organizational contact data
-- GDPR: No personal data — organizational info only (legitimate interest)

CREATE TABLE IF NOT EXISTS club_info (
  id                INTEGER PRIMARY KEY,
  name              TEXT    NOT NULL CHECK(name != ''),
  tagline           TEXT    NOT NULL DEFAULT '',
  founding_year     INTEGER NOT NULL CHECK(founding_year >= 1800 AND founding_year <= 2100),
  short_description TEXT    NOT NULL DEFAULT '' CHECK(length(short_description) <= 300),
  address           TEXT    NOT NULL CHECK(city != ''),
  city              TEXT    NOT NULL CHECK(city != ''),
  postal_code       TEXT    NOT NULL CHECK(postal_code != ''),
  phone             TEXT    NOT NULL CHECK(phone != ''),
  email             TEXT    NOT NULL CHECK(email != '')
) STRICT;

-- Seed row: Alvesta Simsällskap public contact details
-- Fields left blank here MUST be filled via the Trailbase admin UI before go-live
INSERT INTO club_info (
  id, name, tagline, founding_year, short_description,
  address, city, postal_code, phone, email
) VALUES (
  1,
  'Alvesta Simsällskap',
  'Simglädje sedan 1921',
  1921,
  'Simglädje. Hälsa. Gemenskap',
  'Hjortsbergavägen 6C',
  'Alvesta',
  '342 36',
  '076 027 94 10',
  'kansli@alvestass.se'
);
