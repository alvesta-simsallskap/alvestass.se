import type { Instructor, Session, TimeReportConfig } from './types';

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

interface TrailbaseListResponse<T> {
  records: T[];
  cursor: string | null;
}

/**
 * Fetch the single club_info record from Trailbase.
 * Returns `null` if the table is empty (no record yet).
 * Throws on network errors — the caller decides on fallback.
 */
export async function fetchClubInfo(baseUrl: string): Promise<ClubInfo | null> {
  const response = await fetch(
    `${baseUrl}/api/records/v1/club_info?limit=1`
  );

  if (!response.ok) {
    throw new Error(`Trailbase responded with ${response.status}`);
  }

  const body: TrailbaseListResponse<ClubInfo> = await response.json();

  if (!body.records || body.records.length === 0) {
    return null;
  }

  return body.records[0];
}

/**
 * Authenticate as the service user and return the short-lived JWT.
 * Called once at the start of each Worker invocation; the token is reused
 * for all subsequent Trailbase calls within that invocation.
 * Throws on non-2xx responses.
 */
export async function authenticateServiceUser(
  baseUrl: string,
  email: string,
  password: string,
): Promise<string> {
  const response = await fetch(`${baseUrl}/api/auth/v1/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    throw new Error(`Trailbase auth failed with ${response.status}`);
  }

  const body: { auth_token: string } = await response.json();
  return body.auth_token;
}

/**
 * Fetch the single time_report_config row.
 * Returns `null` if the table has no rows.
 * Throws on network errors.
 */
export async function fetchTimeReportConfig(
  baseUrl: string,
  authToken: string,
): Promise<TimeReportConfig | null> {
  const response = await fetch(
    `${baseUrl}/api/records/v1/time_report_config?limit=1`,
    { headers: { Authorization: `Bearer ${authToken}` } },
  );

  if (!response.ok) {
    throw new Error(`Trailbase responded with ${response.status}`);
  }

  const body: TrailbaseListResponse<TimeReportConfig> = await response.json();

  if (!body.records || body.records.length === 0) {
    return null;
  }

  return body.records[0];
}

export interface TrailbaseSession extends Session {
  training_group: string;
}

/**
 * Fetch all sessions for a given month key.
 * Returns an empty array if there are no sessions for that month.
 * Throws on network errors.
 */
export async function fetchTimeReportSessions(
  baseUrl: string,
  monthKey: string,
  authToken: string,
): Promise<TrailbaseSession[]> {
  const url = `${baseUrl}/api/records/v1/time_report_sessions?filter[month_key][$eq]=${encodeURIComponent(monthKey)}&order=date&limit=500`;
  const response = await fetch(url, {
    headers: { Authorization: `Bearer ${authToken}` },
  });

  if (!response.ok) {
    throw new Error(`Trailbase responded with ${response.status}`);
  }

  const body: TrailbaseListResponse<TrailbaseSession> = await response.json();

  return body.records ?? [];
}

/**
 * Fetch a single instructor record by email address.
 * Returns `null` if no instructor is found for that email.
 * Throws on network errors — the caller should catch and omit the salary estimate.
 */
export async function fetchInstructor(
  baseUrl: string,
  email: string,
  authToken: string,
): Promise<Instructor | null> {
  const url = `${baseUrl}/api/records/v1/instructors?filter[email][$eq]=${encodeURIComponent(email)}&limit=1`;
  const response = await fetch(url, {
    headers: { Authorization: `Bearer ${authToken}` },
  });

  if (!response.ok) {
    throw new Error(`Trailbase responded with ${response.status}`);
  }

  const body: TrailbaseListResponse<Instructor> = await response.json();

  if (!body.records || body.records.length === 0) {
    return null;
  }

  return body.records[0];
}
