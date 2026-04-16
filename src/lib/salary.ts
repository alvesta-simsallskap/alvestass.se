// Salary and HTML helpers for time report API
import type { ExtraTimeRow, Instructor, TrainingGroupKey, Session, SessionSchedule, TimeReportConfig } from './types';

const TABLE_ATTRIBUTES = `border=\"1\" cellpadding=\"4\" style=\"border-collapse:collapse;margin-bottom:1em.\"`

export function findTimeItem(schedule: SessionSchedule, group: TrainingGroupKey, value: string): Session | undefined {
  const [date, ...titleParts] = value.split(' ');
  const title = titleParts.join(' ');
  const items = schedule[group] ?? [];
  return items.find(item => item.date === date && item.title === title);
}

export function buildTable(group: TrainingGroupKey, label: string, checked: string[], schedule: SessionSchedule, config: TimeReportConfig) {
  if (!checked.length) return '';
  let rows = '';
  let extraMinutes = 0;

  for (const val of checked) {
    const item = findTimeItem(schedule, group, val);
    let time = '';
    if (item) {
      if (item.hours === 20) time = 'Heldag';
      else if (item.hours === 10) time = 'Halvdag';
      else if (item.hours === 15) time = 'Natt';
      else time = `${item.hours}:${item.minutes < 10 ? '0' : ''}${item.minutes}`;
      rows += `<tr><td>${val}</td><td>${time}</td></tr>`;
      if (group === 'simskola' && item.title === 'Simskola') {
        extraMinutes += config.extra_time_simskola;
      }
      if (item.title === 'Träning') {
        extraMinutes += config.extra_time_training;
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

export function calcSalary(
  group: TrainingGroupKey | 'extratid',
  checked: string[],
  schedule: SessionSchedule,
  config: TimeReportConfig,
  instructor?: Instructor,
  extraRows?: ExtraTimeRow[],
) {
  let hours = 0, minutes = 0;
  let extraMinutes = 0;
  let rate: number | null = null;
  if (!instructor) return { hours, minutes, salary: null, total: 0 };
  if (group === 'simskola') rate = instructor.swim_school_rate;
  else rate = instructor.coach_rate;

  if (group !== 'extratid') {
    for (const val of checked) {
      const item = findTimeItem(schedule, group, val);
      const excluded = new Set([10, 15, 20]);
      if (item && !excluded.has(item.hours)) {
        hours += item.hours;
        minutes += item.minutes;
        if (group === 'simskola' && item.title === 'Simskola') {
          extraMinutes += config.extra_time_simskola;
        }
        if (item.title === 'Träning') {
          extraMinutes += config.extra_time_training;
        }
      }
    }
  }

  // Add extra time rows for extratid section
  if (extraRows && group === 'extratid') {
    for (const row of extraRows) {
      const h = parseInt(row.h) || 0;
      const m = parseInt(row.m) || 0;
      hours += h;
      minutes += m;
    }
  }

  const totalMinutes = hours * 60 + minutes + extraMinutes;
  hours = Math.floor(totalMinutes / 60);
  minutes = totalMinutes % 60;
  const total = rate ? (totalMinutes / 60) * rate : 0;
  return { hours, minutes, salary: rate, total };
}
