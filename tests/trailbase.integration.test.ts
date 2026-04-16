// Integration tests for the live Trailbase instance.
// Skipped automatically unless all three env vars are present.
//
// Run with:
//   TRAILBASE_URL=https://alvestass-trailbase.fly.dev \
//   TRAILBASE_SERVICE_EMAIL=<email> \
//   TRAILBASE_SERVICE_PASSWORD=<password> \
//   pnpm test
//
// Or source your .dev.vars first:
//   export $(grep -v '^#' .dev.vars | xargs) && pnpm test
import { describe, it, expect, beforeAll } from 'vitest';
import {
  authenticateServiceUser,
  fetchTimeReportConfig,
  fetchTimeReportSessions,
  fetchInstructor,
} from '../src/lib/trailbase';

const TRAILBASE_URL = process.env.TRAILBASE_URL;
const TRAILBASE_SERVICE_EMAIL = process.env.TRAILBASE_SERVICE_EMAIL;
const TRAILBASE_SERVICE_PASSWORD = process.env.TRAILBASE_SERVICE_PASSWORD;

const hasCredentials = !!(TRAILBASE_URL && TRAILBASE_SERVICE_EMAIL && TRAILBASE_SERVICE_PASSWORD);

describe.skipIf(!hasCredentials)('Trailbase integration', () => {
  let authToken: string;

  beforeAll(async () => {
    authToken = await authenticateServiceUser(
      TRAILBASE_URL!,
      TRAILBASE_SERVICE_EMAIL!,
      TRAILBASE_SERVICE_PASSWORD!,
    );
  });

  it('authenticates service user and receives a JWT', () => {
    expect(typeof authToken).toBe('string');
    expect(authToken.length).toBeGreaterThan(20);
  });

  it('reads time_report_config (table exists and is accessible)', async () => {
    const config = await fetchTimeReportConfig(TRAILBASE_URL!, authToken);
    expect(config).not.toBeNull();
    expect(config!.active_month_key).toMatch(/^\d{4}-\d{2}$/);
    expect(config!.active_month_display).toBeTruthy();
    console.log(`  active month: ${config!.active_month_key} (${config!.active_month_display})`);
  });

  it('reads time_report_sessions for the active month', async () => {
    const config = await fetchTimeReportConfig(TRAILBASE_URL!, authToken);
    expect(config).not.toBeNull();

    const sessions = await fetchTimeReportSessions(TRAILBASE_URL!, config!.active_month_key, authToken);
    expect(Array.isArray(sessions)).toBe(true);
    console.log(`  sessions for ${config!.active_month_key}: ${sessions.length} found`);

    if (sessions.length > 0) {
      const s = sessions[0];
      expect(s).toHaveProperty('date');
      expect(s).toHaveProperty('title');
      expect(s).toHaveProperty('hours');
      expect(s).toHaveProperty('minutes');
      expect(s).toHaveProperty('training_group');
    }
  });

  it('returns null for an unknown instructor email', async () => {
    const result = await fetchInstructor(TRAILBASE_URL!, 'nobody@example.com', authToken);
    expect(result).toBeNull();
  });
});
