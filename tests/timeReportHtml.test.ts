// Tests for the monetary addon row in src/lib/timeReportHtml.ts
import { describe, it, expect } from 'vitest';
import { buildTimeReportHtml } from '../src/lib/timeReportHtml';
import type { Instructor, SessionSchedule, TimeReportConfig, TimeReportData } from '../src/lib/types';

const config: TimeReportConfig = {
  id: 1,
  active_month_key: '2026-04',
  active_month_display: 'april 2026',
  extra_time_simskola: 30,
  extra_time_training: 15,
  half_day_salary: 500,
  full_day_salary: 1000,
  overnight_salary: 300,
};

const schedule: SessionSchedule = {
  simskola: [{ date: '2026-04-01', title: 'Simskola', hours: 2, minutes: 0 }],
  tavlingA: [],
  tavlingB: [],
  teknik: [],
  masters: [],
  vuxencrawl: [],
};

const baseData: TimeReportData = {
  name: 'Test Testsson',
  email: 'test@example.com',
  milersattning: '',
  kommentarer: '',
  simskola: ['2026-04-01 Simskola'],
  tavlingA: [],
  tavlingB: [],
  teknik: [],
  masters: [],
  vuxencrawl: [],
};

const baseInstructor: Instructor = {
  id: 1,
  email: 'test@example.com',
  name: 'Test Testsson',
  swim_school_rate: 200,
  coach_rate: null,
  travel_compensation: false,
  addon_amount: null,
  addon_description: null,
  fixed_salary: false,
  time_bank: 0,
};

describe('buildTimeReportHtml — monetary addon', () => {
  // simskola 2h + 30 min prep = 2.5h * 200 = 500 kr base
  it('renders an addon row and folds the amount into the total', () => {
    const instructor: Instructor = {
      ...baseInstructor,
      addon_amount: 895,
      addon_description: 'Reseersättning',
    };
    const html = buildTimeReportHtml(baseData, schedule, config, instructor, 'april 2026');

    expect(html).toContain('Reseersättning');
    expect(html).toContain('895,00 kr');
    // base 500 + addon 895 = 1395 (sv-SE uses a narrow no-break space as thousands separator)
    expect(html).toMatch(/1\s395,00 kr/);
  });

  it('renders no addon row when addon fields are null', () => {
    const html = buildTimeReportHtml(baseData, schedule, config, baseInstructor, 'april 2026');
    expect(html).not.toContain('Reseersättning');
    // total equals the base 500 kr
    expect(html).toContain('500,00 kr');
  });

  it('escapes the addon description', () => {
    const instructor: Instructor = {
      ...baseInstructor,
      addon_amount: 100,
      addon_description: '<b>Tillägg</b>',
    };
    const html = buildTimeReportHtml(baseData, schedule, config, instructor, 'april 2026');
    expect(html).toContain('&lt;b&gt;Tillägg&lt;/b&gt;');
    expect(html).not.toContain('<b>Tillägg</b>');
  });
});

describe('buildTimeReportHtml — fixed salary', () => {
  // time_bank 90 min (1:30); simskola 2h + 30 min prep = 150 min worked (2:30)
  const fixedInstructor: Instructor = {
    ...baseInstructor,
    fixed_salary: true,
    time_bank: 90,
  };

  it('shows "Fast lön" and no kr salary table', () => {
    const html = buildTimeReportHtml(baseData, schedule, config, fixedInstructor, 'april 2026');
    expect(html).toContain('Fast lön');
    expect(html).not.toContain('Preliminär löneberäkning');
    expect(html).not.toContain('kr');
  });

  it('renders the time-bank rows (current balance, reported time, estimated new balance)', () => {
    const html = buildTimeReportHtml(baseData, schedule, config, fixedInstructor, 'april 2026');
    expect(html).toContain('Tidbank');
    expect(html).toContain('1:30');       // current balance
    expect(html).toContain('+2:30');      // reported this month
    expect(html).toContain('4:00');       // 90 + 150 = 240 min new balance
  });

  it('renders no monetary addon for a fixed-salary instructor', () => {
    const html = buildTimeReportHtml(baseData, schedule, config, {
      ...fixedInstructor,
      addon_amount: 895,
      addon_description: 'Reseersättning',
    }, 'april 2026');
    expect(html).not.toContain('Reseersättning');
    expect(html).not.toContain('895');
  });

  it('formats a negative time-bank balance', () => {
    const html = buildTimeReportHtml(baseData, schedule, config, {
      ...fixedInstructor,
      time_bank: -60,
    }, 'april 2026');
    expect(html).toContain('-1:00');      // current balance
    // -60 + 150 = 90 min = 1:30 new balance
    expect(html).toContain('1:30');
  });
});
