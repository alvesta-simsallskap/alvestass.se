// Tests for src/lib/salary.ts — updated for refactored pure-function signatures
import { describe, it, expect } from 'vitest';
import { findTimeItem, buildTable, calcSalary } from '../src/lib/salary';
import type { Instructor, SessionSchedule, TimeReportConfig } from '../src/lib/types';

// ─── Fixtures ────────────────────────────────────────────────────────────────

const defaultConfig: TimeReportConfig = {
  id: 1,
  active_month_key: '2026-04',
  active_month_display: 'april 2026',
  extra_time_simskola: 30,
  extra_time_training: 15,
  half_day_salary: 500,
  full_day_salary: 1000,
  overnight_salary: 300,
};

const defaultSchedule: SessionSchedule = {
  simskola: [
    { date: '2026-04-01', title: 'Simskola', hours: 5, minutes: 10 },
    { date: '2026-04-01', title: 'Beredskap', hours: 3, minutes: 0 },
    { date: '2026-04-08', title: 'Simskola', hours: 5, minutes: 10 },
  ],
  tavlingA: [
    { date: '2026-04-02', title: 'Träning', hours: 1, minutes: 45 },
    { date: '2026-04-04', title: 'Träning', hours: 1, minutes: 45 },
    { date: '2026-04-14', title: 'ÖGP 2 Kalmar', hours: 20, minutes: 0 },
    { date: '2026-04-14', title: 'ÖGP 2 Övernattning', hours: 15, minutes: 0 },
    { date: '2026-04-15', title: 'Halvdag Linköping', hours: 10, minutes: 0 },
  ],
  tavlingB: [],
  teknik: [],
  masters: [
    { date: '2026-04-03', title: 'Träning', hours: 1, minutes: 30 },
  ],
  vuxencrawl: [],
};

const instructorFull: Instructor = {
  id: 1,
  email: 'test@example.com',
  name: 'Test Testsson',
  swim_school_rate: 200,
  coach_rate: 150,
  travel_compensation: false,
  addon_amount: null,
  addon_description: null,
};

const instructorNoCoach: Instructor = {
  id: 2,
  email: 'nocoach@example.com',
  name: 'Ingen Coach',
  swim_school_rate: 180,
  coach_rate: null,
  travel_compensation: false,
  addon_amount: null,
  addon_description: null,
};

// ─── findTimeItem ─────────────────────────────────────────────────────────────

describe('findTimeItem', () => {
  it('finds a simskola session by date and title', () => {
    const item = findTimeItem(defaultSchedule, 'simskola', '2026-04-01 Simskola');
    expect(item).toBeDefined();
    expect(item?.date).toBe('2026-04-01');
    expect(item?.title).toBe('Simskola');
    expect(item?.hours).toBe(5);
    expect(item?.minutes).toBe(10);
  });

  it('finds a simskola session with a single-word title', () => {
    const item = findTimeItem(defaultSchedule, 'simskola', '2026-04-01 Beredskap');
    expect(item).toBeDefined();
    expect(item?.title).toBe('Beredskap');
    expect(item?.hours).toBe(3);
    expect(item?.minutes).toBe(0);
  });

  it('finds a full-day competition session (hours=20)', () => {
    const item = findTimeItem(defaultSchedule, 'tavlingA', '2026-04-14 ÖGP 2 Kalmar');
    expect(item).toBeDefined();
    expect(item?.hours).toBe(20);
    expect(item?.minutes).toBe(0);
  });

  it('finds an overnight session (hours=15)', () => {
    const item = findTimeItem(defaultSchedule, 'tavlingA', '2026-04-14 ÖGP 2 Övernattning');
    expect(item).toBeDefined();
    expect(item?.hours).toBe(15);
  });

  it('finds a half-day session (hours=10)', () => {
    const item = findTimeItem(defaultSchedule, 'tavlingA', '2026-04-15 Halvdag Linköping');
    expect(item).toBeDefined();
    expect(item?.hours).toBe(10);
  });

  it('returns undefined for a non-existent session', () => {
    const item = findTimeItem(defaultSchedule, 'simskola', '2026-04-01 Nonexistent Activity');
    expect(item).toBeUndefined();
  });

  it('returns undefined for an empty value string', () => {
    const item = findTimeItem(defaultSchedule, 'simskola', '');
    expect(item).toBeUndefined();
  });

  it('returns undefined for an empty group', () => {
    const item = findTimeItem(defaultSchedule, 'tavlingB', '2026-04-01 Simskola');
    expect(item).toBeUndefined();
  });
});

// ─── buildTable ──────────────────────────────────────────────────────────────

describe('buildTable', () => {
  it('returns empty string when checked array is empty', () => {
    expect(buildTable('simskola', 'Simskola', [], defaultSchedule, defaultConfig)).toBe('');
  });

  it('builds an HTML table for a known simskola session', () => {
    const html = buildTable('simskola', 'Simskola', ['2026-04-01 Simskola'], defaultSchedule, defaultConfig);
    expect(html).toContain('<h4>Simskola</h4>');
    expect(html).toContain('<table');
    expect(html).toContain('2026-04-01 Simskola');
    // Simskola title adds 30 min extra
    expect(html).toContain('Föreberedelser och undanplockning');
    expect(html).toContain('0:30');
  });

  it('builds an HTML table for a training session with 15 min prep time', () => {
    const html = buildTable('tavlingA', 'Tävling A', ['2026-04-02 Träning'], defaultSchedule, defaultConfig);
    expect(html).toContain('<h4>Tävling A</h4>');
    expect(html).toContain('Föreberedelser och undanplockning');
    expect(html).toContain('0:15');
  });

  it('shows "Heldag" label for full-day competition (hours=20)', () => {
    const html = buildTable('tavlingA', 'Tävling A', ['2026-04-14 ÖGP 2 Kalmar'], defaultSchedule, defaultConfig);
    expect(html).toContain('Heldag');
  });

  it('shows "Halvdag" label for half-day competition (hours=10)', () => {
    const html = buildTable('tavlingA', 'Tävling A', ['2026-04-15 Halvdag Linköping'], defaultSchedule, defaultConfig);
    expect(html).toContain('Halvdag');
  });

  it('shows "Natt" label for overnight session (hours=15)', () => {
    const html = buildTable('tavlingA', 'Tävling A', ['2026-04-14 ÖGP 2 Övernattning'], defaultSchedule, defaultConfig);
    expect(html).toContain('Natt');
  });

  it('renders empty time cell for unknown session', () => {
    const html = buildTable('simskola', 'Simskola', ['2026-04-99 Nonexistent'], defaultSchedule, defaultConfig);
    expect(html).toContain('2026-04-99 Nonexistent');
    expect(html).not.toContain('Föreberedelser');
  });

  it('accumulates extra time across multiple simskola sessions', () => {
    const html = buildTable('simskola', 'Simskola', [
      '2026-04-01 Simskola',
      '2026-04-08 Simskola',
    ], defaultSchedule, defaultConfig);
    // 2 × 30 min = 60 min = 1:00
    expect(html).toContain('1:00');
  });

  it('uses config values for extra time (custom config)', () => {
    const customConfig: TimeReportConfig = { ...defaultConfig, extra_time_simskola: 45 };
    const html = buildTable('simskola', 'Simskola', ['2026-04-01 Simskola'], defaultSchedule, customConfig);
    expect(html).toContain('0:45');
  });
});

// ─── calcSalary ──────────────────────────────────────────────────────────────

describe('calcSalary', () => {
  it('returns zero totals and null salary when no instructor given', () => {
    const result = calcSalary('simskola', ['2026-04-01 Simskola'], defaultSchedule, defaultConfig);
    expect(result.hours).toBe(0);
    expect(result.minutes).toBe(0);
    expect(result.salary).toBeNull();
    expect(result.total).toBe(0);
  });

  it('calculates simskola salary using swim_school_rate', () => {
    const result = calcSalary('simskola', ['2026-04-01 Simskola'], defaultSchedule, defaultConfig, instructorFull);
    // 5h 10min + 30min prep = 5h 40min = 340 min
    expect(result.salary).toBe(200);
    expect(result.hours).toBe(5);
    expect(result.minutes).toBe(40);
    expect(result.total).toBeCloseTo((340 / 60) * 200, 4);
  });

  it('calculates coaching salary using coach_rate', () => {
    const result = calcSalary('tavlingA', ['2026-04-02 Träning'], defaultSchedule, defaultConfig, instructorFull);
    // 1h 45min + 15min prep = 2h 0min = 120 min
    expect(result.salary).toBe(150);
    expect(result.hours).toBe(2);
    expect(result.minutes).toBe(0);
    expect(result.total).toBeCloseTo(2 * 150, 4);
  });

  it('returns total=0 when coach_rate is null', () => {
    const result = calcSalary('tavlingA', ['2026-04-02 Träning'], defaultSchedule, defaultConfig, instructorNoCoach);
    expect(result.salary).toBeNull();
    expect(result.total).toBe(0);
  });

  it('excludes full-day competition (hours=20) from hourly calculation', () => {
    const result = calcSalary('tavlingA', ['2026-04-14 ÖGP 2 Kalmar'], defaultSchedule, defaultConfig, instructorFull);
    expect(result.hours).toBe(0);
    expect(result.minutes).toBe(0);
    expect(result.total).toBe(0);
  });

  it('excludes overnight session (hours=15) from hourly calculation', () => {
    const result = calcSalary('tavlingA', ['2026-04-14 ÖGP 2 Övernattning'], defaultSchedule, defaultConfig, instructorFull);
    expect(result.hours).toBe(0);
    expect(result.total).toBe(0);
  });

  it('excludes half-day session (hours=10) from hourly calculation', () => {
    const result = calcSalary('tavlingA', ['2026-04-15 Halvdag Linköping'], defaultSchedule, defaultConfig, instructorFull);
    expect(result.hours).toBe(0);
    expect(result.total).toBe(0);
  });

  it('returns zero result for empty checked array with instructor', () => {
    const result = calcSalary('simskola', [], defaultSchedule, defaultConfig, instructorFull);
    expect(result.hours).toBe(0);
    expect(result.minutes).toBe(0);
    expect(result.total).toBe(0);
  });

  it('sums extra time rows for extratid section', () => {
    const extraRows = [
      { date: '2026-04-10', h: '1', m: '30', desc: 'Möte' },
      { date: '2026-04-11', h: '0', m: '45', desc: 'Admin' },
    ];
    const result = calcSalary('extratid', [], defaultSchedule, defaultConfig, instructorFull, extraRows);
    // 1h30 + 0h45 = 2h15 = 135 min
    expect(result.hours).toBe(2);
    expect(result.minutes).toBe(15);
    // extratid uses coach_rate (not simskola rate)
    expect(result.salary).toBe(150);
    expect(result.total).toBeCloseTo((135 / 60) * 150, 4);
  });

  it('uses config values for extra time calculation (custom config)', () => {
    const customConfig: TimeReportConfig = { ...defaultConfig, extra_time_training: 20 };
    const result = calcSalary('tavlingA', ['2026-04-02 Träning'], defaultSchedule, customConfig, instructorFull);
    // 1h 45min + 20min prep = 2h 5min = 125 min
    expect(result.hours).toBe(2);
    expect(result.minutes).toBe(5);
  });
});
