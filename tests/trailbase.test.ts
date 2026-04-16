// Unit tests for src/lib/trailbase.ts — fetch is mocked; no real network calls.
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  authenticateServiceUser,
  fetchTimeReportConfig,
  fetchTimeReportSessions,
  fetchInstructor,
} from '../src/lib/trailbase';

// ─── Mock setup ──────────────────────────────────────────────────────────────

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

function mockOk(body: unknown) {
  return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve(body) });
}

function mockError(status: number) {
  return Promise.resolve({ ok: false, status, json: () => Promise.resolve({}) });
}

beforeEach(() => mockFetch.mockReset());

// ─── authenticateServiceUser ──────────────────────────────────────────────────

describe('authenticateServiceUser', () => {
  it('POSTs to /api/auth/v1/login and returns auth_token', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ auth_token: 'jwt-abc', refresh_token: 'r', csrf_token: 'c' }));

    const token = await authenticateServiceUser('https://tb.example.com', 'svc@example.com', 'secret');

    expect(token).toBe('jwt-abc');
    expect(mockFetch).toHaveBeenCalledWith(
      'https://tb.example.com/api/auth/v1/login',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: 'svc@example.com', password: 'secret' }),
      }),
    );
  });

  it('throws on 401 (bad credentials)', async () => {
    mockFetch.mockReturnValueOnce(mockError(401));
    await expect(
      authenticateServiceUser('https://tb.example.com', 'bad@example.com', 'wrong'),
    ).rejects.toThrow('401');
  });

  it('throws on network error', async () => {
    mockFetch.mockRejectedValueOnce(new Error('fetch failed'));
    await expect(
      authenticateServiceUser('https://tb.example.com', 'svc@example.com', 'secret'),
    ).rejects.toThrow('fetch failed');
  });
});

// ─── fetchTimeReportConfig ────────────────────────────────────────────────────

describe('fetchTimeReportConfig', () => {
  const configRecord = {
    id: 1,
    active_month_key: '2026-04',
    active_month_display: 'april 2026',
    extra_time_simskola: 30,
    extra_time_training: 15,
    half_day_salary: 500,
    full_day_salary: 1000,
    overnight_salary: 300,
  };

  it('sends Authorization header and returns the first config record', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [configRecord], cursor: null }));

    const result = await fetchTimeReportConfig('https://tb.example.com', 'my-jwt');

    expect(result).toEqual(configRecord);
    expect(mockFetch).toHaveBeenCalledWith(
      'https://tb.example.com/api/records/v1/time_report_config?limit=1',
      expect.objectContaining({ headers: { Authorization: 'Bearer my-jwt' } }),
    );
  });

  it('returns null when records array is empty', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [], cursor: null }));
    const result = await fetchTimeReportConfig('https://tb.example.com', 'my-jwt');
    expect(result).toBeNull();
  });

  it('throws on 403 (access policy not set)', async () => {
    mockFetch.mockReturnValueOnce(mockError(403));
    await expect(fetchTimeReportConfig('https://tb.example.com', 'my-jwt')).rejects.toThrow('403');
  });

  it('throws on 401 (token expired)', async () => {
    mockFetch.mockReturnValueOnce(mockError(401));
    await expect(fetchTimeReportConfig('https://tb.example.com', 'my-jwt')).rejects.toThrow('401');
  });
});

// ─── fetchTimeReportSessions ──────────────────────────────────────────────────

describe('fetchTimeReportSessions', () => {
  const sessionRecord = {
    id: 1,
    month_key: '2026-04',
    training_group: 'simskola',
    date: '2026-04-01',
    title: 'Simskola',
    hours: 5,
    minutes: 10,
  };

  it('sends correct filter URL and Authorization header', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [sessionRecord], cursor: null }));

    const results = await fetchTimeReportSessions('https://tb.example.com', '2026-04', 'my-jwt');

    expect(results).toHaveLength(1);
    expect(results[0].title).toBe('Simskola');
    expect(results[0].hours).toBe(5);
    const calledUrl: string = mockFetch.mock.calls[0][0];
    expect(calledUrl).toContain('filter[month_key][$eq]=2026-04');
    expect(calledUrl).toContain('limit=500');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ headers: { Authorization: 'Bearer my-jwt' } }),
    );
  });

  it('URL-encodes the month key', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [], cursor: null }));
    await fetchTimeReportSessions('https://tb.example.com', '2026-04', 'tok');
    const calledUrl: string = mockFetch.mock.calls[0][0];
    // encodeURIComponent('2026-04') = '2026-04' (hyphen is not encoded, but verify no spaces/special chars break it)
    expect(calledUrl).toContain('2026-04');
  });

  it('returns empty array when no sessions exist for the month', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [], cursor: null }));
    const results = await fetchTimeReportSessions('https://tb.example.com', '2026-05', 'my-jwt');
    expect(results).toEqual([]);
  });

  it('throws on 403 (access policy not set)', async () => {
    mockFetch.mockReturnValueOnce(mockError(403));
    await expect(
      fetchTimeReportSessions('https://tb.example.com', '2026-04', 'my-jwt'),
    ).rejects.toThrow('403');
  });
});

// ─── fetchInstructor ─────────────────────────────────────────────────────────

describe('fetchInstructor', () => {
  const instructorRecord = {
    id: 7,
    email: 'coach@example.com',
    swim_school_rate: 115,
    coach_rate: 145,
  };

  it('sends email filter and Authorization header, returns instructor', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [instructorRecord], cursor: null }));

    const result = await fetchInstructor('https://tb.example.com', 'coach@example.com', 'my-jwt');

    expect(result).toEqual(instructorRecord);
    const calledUrl: string = mockFetch.mock.calls[0][0];
    expect(calledUrl).toContain('filter[email][$eq]=');
    expect(calledUrl).toContain('coach%40example.com'); // @ is encoded
    expect(calledUrl).toContain('limit=1');
  });

  it('URL-encodes the email address', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [], cursor: null }));
    await fetchInstructor('https://tb.example.com', 'user+tag@example.com', 'tok');
    const calledUrl: string = mockFetch.mock.calls[0][0];
    expect(calledUrl).toContain(encodeURIComponent('user+tag@example.com'));
  });

  it('returns null when email is not found', async () => {
    mockFetch.mockReturnValueOnce(mockOk({ records: [], cursor: null }));
    const result = await fetchInstructor('https://tb.example.com', 'unknown@example.com', 'my-jwt');
    expect(result).toBeNull();
  });

  it('throws on 403 (instructors table not readable by service user)', async () => {
    mockFetch.mockReturnValueOnce(mockError(403));
    await expect(
      fetchInstructor('https://tb.example.com', 'coach@example.com', 'my-jwt'),
    ).rejects.toThrow('403');
  });
});
