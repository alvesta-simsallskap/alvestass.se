// Validation and type helpers for time report API (Cloudflare/Edge compatible)
import type { TimeReportData, ExtraTimeRow } from './types';

export function parseTimeReportForm(formData: FormData): TimeReportData {
  // Parse extra time rows
  const extraRows: ExtraTimeRow[] = [];
  let idx = 0;
  while (true) {
    const date = formData.get(`extratid[${idx}][date]`);
    const h = formData.get(`extratid[${idx}][h]`);
    const m = formData.get(`extratid[${idx}][m]`);
    const desc = formData.get(`extratid[${idx}][desc]`);
    if (date || h || m || desc) {
      if (date && desc) {
        const hours = String(h || '0');
        const minutes = String(m || '0');
        extraRows.push({ date: String(date), h: hours, m: minutes, desc: String(desc) });
      }
      idx++;
    } else {
      break;
    }
  }
  return {
    name: String(formData.get('namn') || ''),
    email: String(formData.get('email') || '').toLowerCase().trim(),
    milersattning: String(formData.get('milersattning') || ''),
    kommentarer: String(formData.get('kommentarer') || ''),
    simskola: formData.getAll('simskola_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    tavlingA: formData.getAll('tavling-a_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    tavlingB: formData.getAll('tavling-b_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    teknik: formData.getAll('teknik_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    masters: formData.getAll('masters_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    vuxencrawl: formData.getAll('vuxencrawl_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    extratid: extraRows,
  };
}

// Add further validation as needed (e.g. Zod, custom checks)
