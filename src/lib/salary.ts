// Salary and HTML helpers for time report API
import type { Employee, TimeReportData, ExtraTimeRow } from './types';
import timeReportItems from '../config/time-report-items.json';
import { EXTRA_TIME_SIMSKOLA, EXTRA_TIME_TRAINING } from '../config/time-report-settings';

const TABLE_ATTRIBUTES = `border=\"1\" cellpadding=\"4\" style=\"border-collapse:collapse;margin-bottom:1em;\"`

export function findTimeItem(section: string, value: string) {
  const [date, ...titleParts] = value.split(' ');
  const title = titleParts.join(' ');
  const items = (timeReportItems['2026-02'] as any)[section] || [];
  return items.find((item: any) => item.date === date && item.title === title);
}

export function buildTable(section: string, label: string, checked: string[]) {
  if (!checked.length) return '';
  let rows = '';
  let extraMinutes = 0;
  
  for (const val of checked) {
    const item = findTimeItem(section, val);
    let time = '';
    if (item) {
      if (item.h === 20) time = 'Heldag';
      else if (item.h === 10) time = 'Halvdag';
      else time = `${item.h}:${item.m < 10 ? '0' : ''}${item.m}`;
      rows += `<tr><td>${val}</td><td>${time}</td></tr>`;
      // Sum up extra time for summary row
      if (section === 'simskola' && item.title === 'Simskola') {
        extraMinutes += EXTRA_TIME_SIMSKOLA;
      }
      if (item.title === 'Träning') {
        extraMinutes += EXTRA_TIME_TRAINING;
      }
    } else {
      rows += `<tr><td>${val}</td><td></td></tr>`;
    }
  }

  if (extraMinutes > 0) {
    const extraH = Math.floor(extraMinutes / 60);
    const extraM = extraMinutes % 60;
    rows += `<tr><td>Föreberedelser och undanplockning</td><td>${extraH}:${extraM < 10 ? '0' : ''}${extraM}</td></tr>`;
  }

  return `<h4>${label}</h4><table ${TABLE_ATTRIBUTES}><thead><tr><th>Datum och aktivitet</th><th>Tid</th></tr></thead><tbody>${rows}</tbody></table>`;
}

export function buildAdditionalTimeTable(extraRows: ExtraTimeRow[]) {
  if (!extraRows.length) return '';
  
  let rows = '';
  
  for (const row of extraRows) {
    rows += `<tr><td>${row.date} ${row.desc}</td><td>${row.h}:${parseInt(row.m) < 10 ? '0' : ''}${row.m}</td></tr>`;
  }

  return `<h4>Övrig tid</h4><table ${TABLE_ATTRIBUTES}><thead><tr><th>Beskrivning</th><th>Tid</th></tr></thead><tbody>${rows}</tbody></table>`;
}

export function calcSalary(section: string, checked: string[], employee?: Employee, extraRows?: ExtraTimeRow[]) {
  let hours = 0, minutes = 0;
  let extraMinutes = 0;
  let rate: number|null = null;
  if (!employee) return { hours, minutes, salary: null, total: 0 };
  if (section === 'simskola') rate = employee.swimSchoolRate;
  else rate = employee.coachRate;
  
  for (const val of checked) {
    const item = findTimeItem(section, val);
    const excluded = new Set([10, 20]);
    if (item && !excluded.has(item.h)) {
      hours += item.h;
      minutes += item.m;
      if (section === 'simskola' && item.title === 'Simskola') {
        extraMinutes += EXTRA_TIME_SIMSKOLA;
      }
      if (item.title === 'Träning') {
        extraMinutes += EXTRA_TIME_TRAINING;
      }
    }
  }

  // Add extra time rows if provided (for coach rate)
  if (extraRows && section === 'extratid') {
    for (const row of extraRows) {
      const h = parseInt(row.h) || 0;
      const m = parseInt(row.m) || 0;
      hours += h;
      minutes += m;
    }
  }

  let totalMinutes = hours * 60 + minutes + extraMinutes;
  hours = Math.floor(totalMinutes / 60);
  minutes = totalMinutes % 60;
  const total = rate ? (totalMinutes / 60) * rate : 0;
  return { hours, minutes, salary: rate, total };
}
