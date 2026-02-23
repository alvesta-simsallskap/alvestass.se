// Validation and type helpers for time report API (Cloudflare/Edge compatible)
import type { TimeReportData } from './types';

export function parseTimeReportForm(formData: FormData): TimeReportData {
  return {
    name: String(formData.get('namn') || ''),
    email: String(formData.get('email') || ''),
    milersattning: String(formData.get('milersattning') || ''),
    kommentarer: String(formData.get('kommentarer') || ''),
    simskola: formData.getAll('simskola_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    tavlingA: formData.getAll('tavling-a_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    tavlingB: formData.getAll('tavling-b_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    teknik: formData.getAll('teknik_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    masters: formData.getAll('masters_checked_dates[]').filter((v): v is string => typeof v === 'string'),
    vuxencrawl: formData.getAll('vuxencrawl_checked_dates[]').filter((v): v is string => typeof v === 'string'),
  };
}

// Add further validation as needed (e.g. Zod, custom checks)
