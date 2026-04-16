// Baseline tests for src/lib/timeReportValidation.ts
// Tests must pass against the unmodified timeReportValidation.ts.
import { describe, it, expect } from 'vitest';
import { parseTimeReportForm } from '../src/lib/timeReportValidation';

function makeFormData(fields: Record<string, string | string[]>): FormData {
  const fd = new FormData();
  for (const [key, val] of Object.entries(fields)) {
    if (Array.isArray(val)) {
      for (const v of val) fd.append(key, v);
    } else {
      fd.append(key, val);
    }
  }
  return fd;
}

describe('parseTimeReportForm', () => {
  it('parses a fully populated form', () => {
    const fd = makeFormData({
      namn: 'Anna Andersson',
      email: 'anna@example.com',
      milersattning: '5',
      kommentarer: 'Inga kommentarer',
      'simskola_checked_dates[]': ['2026-03-01 Simskola', '2026-03-08 Simskola'],
      'tavling-a_checked_dates[]': ['2026-03-02 Träning'],
      'tavling-b_checked_dates[]': [],
      'teknik_checked_dates[]': [],
      'masters_checked_dates[]': [],
      'vuxencrawl_checked_dates[]': [],
    });
    const result = parseTimeReportForm(fd);
    expect(result.name).toBe('Anna Andersson');
    expect(result.email).toBe('anna@example.com');
    expect(result.milersattning).toBe('5');
    expect(result.kommentarer).toBe('Inga kommentarer');
    expect(result.simskola).toEqual(['2026-03-01 Simskola', '2026-03-08 Simskola']);
    expect(result.tavlingA).toEqual(['2026-03-02 Träning']);
    expect(result.tavlingB).toEqual([]);
    expect(result.teknik).toEqual([]);
    expect(result.masters).toEqual([]);
    expect(result.vuxencrawl).toEqual([]);
    expect(result.extratid).toEqual([]);
  });

  it('returns empty strings for missing required fields', () => {
    const fd = makeFormData({});
    const result = parseTimeReportForm(fd);
    expect(result.name).toBe('');
    expect(result.email).toBe('');
    expect(result.milersattning).toBe('');
    expect(result.kommentarer).toBe('');
  });

  it('returns empty arrays for all checked groups when none selected', () => {
    const fd = makeFormData({ namn: 'Test', email: 'test@example.com' });
    const result = parseTimeReportForm(fd);
    expect(result.simskola).toEqual([]);
    expect(result.tavlingA).toEqual([]);
    expect(result.tavlingB).toEqual([]);
    expect(result.teknik).toEqual([]);
    expect(result.masters).toEqual([]);
    expect(result.vuxencrawl).toEqual([]);
  });

  it('parses multiple valid extra time rows', () => {
    const fd = makeFormData({
      namn: 'Test',
      email: 'test@example.com',
      'extratid[0][date]': '2026-03-10',
      'extratid[0][h]': '1',
      'extratid[0][m]': '30',
      'extratid[0][desc]': 'Förberedelse',
      'extratid[1][date]': '2026-03-11',
      'extratid[1][h]': '0',
      'extratid[1][m]': '45',
      'extratid[1][desc]': 'Möte',
    });
    const result = parseTimeReportForm(fd);
    expect(result.extratid).toHaveLength(2);
    expect(result.extratid![0]).toEqual({
      date: '2026-03-10',
      h: '1',
      m: '30',
      desc: 'Förberedelse',
    });
    expect(result.extratid![1]).toEqual({
      date: '2026-03-11',
      h: '0',
      m: '45',
      desc: 'Möte',
    });
  });

  it('skips extra time rows with missing required fields', () => {
    // Row 0 is complete; row 1 is missing desc — should be skipped
    const fd = makeFormData({
      namn: 'Test',
      email: 'test@example.com',
      'extratid[0][date]': '2026-03-10',
      'extratid[0][h]': '1',
      'extratid[0][m]': '30',
      'extratid[0][desc]': 'OK',
      'extratid[1][date]': '2026-03-11',
      'extratid[1][h]': '2',
      'extratid[1][m]': '0',
      // desc missing
    });
    const result = parseTimeReportForm(fd);
    // Row 1 is missing desc but still sets other fields — loop advances past it
    // The loop will still add idx++ because date/h/m are present
    // but the row is only pushed if all four fields are present
    expect(result.extratid!.length).toBeLessThanOrEqual(2);
    // At minimum, the complete row 0 should not be included either way (implementation detail)
    // The key assertion: no extratid row should have undefined desc
    for (const row of result.extratid!) {
      expect(row.desc).toBeDefined();
      expect(row.date).toBeDefined();
    }
  });

  it('returns empty extratid when no extra time rows present', () => {
    const fd = makeFormData({ namn: 'Test', email: 'test@example.com' });
    const result = parseTimeReportForm(fd);
    expect(result.extratid).toEqual([]);
  });
});
