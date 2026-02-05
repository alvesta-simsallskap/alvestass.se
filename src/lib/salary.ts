// Salary and HTML helpers for time report API
import type { Employee, TimeReportData } from './types';
import timeReportItems from '../config/time-report-items.json';

export function findTimeItem(section: string, value: string) {
  const [date, ...titleParts] = value.split(' ');
  const title = titleParts.join(' ');
  const items = (timeReportItems['2026-01'] as any)[section] || [];
  return items.find((item: any) => item.date === date && item.title === title);
}

export function buildTable(section: string, label: string, checked: string[]) {
  if (!checked.length) return '';
  let rows = '';
  for (const val of checked) {
    const item = findTimeItem(section, val);
    let time = '';
    if (item) {
      if (item.h === 20) time = 'Heldag';
      else if (item.h === 10) time = 'Halvdag';
      else time = `${item.h}:${item.m < 10 ? '0' : ''}${item.m}`;
      rows += `<tr><td>${val}</td><td>${time}</td></tr>`;
    } else {
      rows += `<tr><td>${val}</td><td></td></tr>`;
    }
  }
  return `<h4>${label}</h4><table border=\"1\" cellpadding=\"4\" style=\"border-collapse:collapse;margin-bottom:1em;\"><thead><tr><th>Datum och aktivitet</th><th>Tid</th></tr></thead><tbody>${rows}</tbody></table>`;
}

export function calcSalary(section: string, checked: string[], employee?: Employee) {
  let hours = 0, minutes = 0;
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
    }
  }
  let totalMinutes = hours * 60 + minutes;
  hours = Math.floor(totalMinutes / 60);
  minutes = totalMinutes % 60;
  const total = rate ? (totalMinutes / 60) * rate : 0;
  return { hours, minutes, salary: rate, total };
}
