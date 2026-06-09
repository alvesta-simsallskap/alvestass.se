import type { TimeReportData, SessionSchedule, TimeReportConfig, Instructor } from './types';
import { buildTable, buildAdditionalTimeTable, calcSalary, findTimeItem } from './salary';

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

export interface AttachmentInfo {
  filename: string;
  desc?: string;
}

export function buildTimeReportHtml(
  data: TimeReportData,
  schedule: SessionSchedule,
  config: TimeReportConfig,
  instructor: Instructor | undefined,
  activeMonthDisplay: string,
  attachments?: AttachmentInfo[],
): string {
  const formatAmount = (amount: number) => amount.toLocaleString('sv-SE', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

  let html = `<h4>Tidrapport ${escapeHtml(activeMonthDisplay)}</h4>`;
  html += `<p><b>Namn:</b> ${escapeHtml(data.name)}</p>`;

  // Session groups
  html += buildTable('simskola', 'Simskola', data.simskola, schedule, config);
  html += buildTable('tavlingA', 'Tävlingsgrupp A', data.tavlingA, schedule, config);
  html += buildTable('tavlingB', 'Tävlingsgrupp B', data.tavlingB, schedule, config);
  html += buildTable('teknik', 'Teknik', data.teknik, schedule, config);
  html += buildTable('masters', 'Masters', data.masters, schedule, config);
  html += buildTable('vuxencrawl', 'Vuxencrawl', data.vuxencrawl, schedule, config);

  // Extra time table
  if (data.extratid) {
    html += buildAdditionalTimeTable(data.extratid);
  }

  // Mileage
  if (data.milersattning) {
    html += `<p><b>Milersättning:</b> ${escapeHtml(data.milersattning)} km (privat resa, skattepliktigt, 25 kr/mil)</p>`;
  }

  // Comments
  if (data.kommentarer) {
    html += `<p><b>Kommentarer:</b> ${escapeHtml(data.kommentarer)}</p>`;
  }

  // Salary estimate (only if instructor found)
  if (instructor) {
    const salarySimskola = calcSalary('simskola', data.simskola, schedule, config, instructor);
    const salaryTavlingA = calcSalary('tavlingA', data.tavlingA, schedule, config, instructor);
    const salaryTavlingB = calcSalary('tavlingB', data.tavlingB, schedule, config, instructor);
    const salaryTeknik = calcSalary('teknik', data.teknik, schedule, config, instructor);
    const salaryMasters = calcSalary('masters', data.masters, schedule, config, instructor);
    const salaryVuxencrawl = calcSalary('vuxencrawl', data.vuxencrawl, schedule, config, instructor);
    const salaryOvrigTid = calcSalary('extratid', [], schedule, config, instructor, data.extratid);

    // Count flat-rate competition sessions
    let fullDay = 0, halfDay = 0, overnight = 0;
    const countSpecial = (group: 'tavlingA' | 'tavlingB' | 'masters' | 'teknik', vals: string[]) => {
      for (const val of vals) {
        const item = findTimeItem(schedule, group, val);
        if (item?.hours === 20) fullDay++;
        if (item?.hours === 10) halfDay++;
        if (item?.hours === 15) overnight++;
      }
    };
    countSpecial('tavlingA', data.tavlingA);
    countSpecial('tavlingB', data.tavlingB);
    countSpecial('masters', data.masters);
    countSpecial('teknik', data.teknik);

    const fullDaySalary = config.full_day_salary * fullDay;
    const halfDaySalary = config.half_day_salary * halfDay;
    const overnightSalary = config.overnight_salary * overnight;
    const totalSalary = [salarySimskola, salaryTavlingA, salaryTavlingB, salaryTeknik, salaryMasters, salaryVuxencrawl, salaryOvrigTid]
      .reduce((sum, s) => sum + s.total, 0);

    html += `<h4>Preliminär löneberäkning</h4><table border="1" cellpadding="4" style="border-collapse:collapse;margin-bottom:1em;">
      <thead><tr><th>Grupp</th><th>Timmar</th><th>Minuter</th><th>Lön</th><th>Summa</th></tr></thead><tbody>`;

    if (salarySimskola.hours > 0 || salarySimskola.minutes > 0) {
      html += `<tr><td>Simskola</td><td>${salarySimskola.hours}</td><td>${salarySimskola.minutes}</td><td>${salarySimskola.salary ?? '-'}</td><td>${formatAmount(salarySimskola.total)} kr</td></tr>`;
    }
    if (salaryTavlingA.hours > 0 || salaryTavlingA.minutes > 0) {
      html += `<tr><td>Tävlingsgrupp A</td><td>${salaryTavlingA.hours}</td><td>${salaryTavlingA.minutes}</td><td>${salaryTavlingA.salary ?? '-'}</td><td>${formatAmount(salaryTavlingA.total)} kr</td></tr>`;
    }
    if (salaryTavlingB.hours > 0 || salaryTavlingB.minutes > 0) {
      html += `<tr><td>Tävlingsgrupp B</td><td>${salaryTavlingB.hours}</td><td>${salaryTavlingB.minutes}</td><td>${salaryTavlingB.salary ?? '-'}</td><td>${formatAmount(salaryTavlingB.total)} kr</td></tr>`;
    }
    if (salaryTeknik.hours > 0 || salaryTeknik.minutes > 0) {
      html += `<tr><td>Teknik</td><td>${salaryTeknik.hours}</td><td>${salaryTeknik.minutes}</td><td>${salaryTeknik.salary ?? '-'}</td><td>${formatAmount(salaryTeknik.total)} kr</td></tr>`;
    }
    if (salaryMasters.hours > 0 || salaryMasters.minutes > 0) {
      html += `<tr><td>Masters</td><td>${salaryMasters.hours}</td><td>${salaryMasters.minutes}</td><td>${salaryMasters.salary ?? '-'}</td><td>${formatAmount(salaryMasters.total)} kr</td></tr>`;
    }
    if (salaryVuxencrawl.hours > 0 || salaryVuxencrawl.minutes > 0) {
      html += `<tr><td>Vuxencrawl</td><td>${salaryVuxencrawl.hours}</td><td>${salaryVuxencrawl.minutes}</td><td>${salaryVuxencrawl.salary ?? '-'}</td><td>${formatAmount(salaryVuxencrawl.total)} kr</td></tr>`;
    }
    if (fullDay > 0) {
      html += `<tr><td>Heldagar</td><td colspan="2">${fullDay}</td><td>${config.full_day_salary}</td><td>${formatAmount(fullDaySalary)} kr</td></tr>`;
    }
    if (halfDay > 0) {
      html += `<tr><td>Halvdagar</td><td colspan="2">${halfDay}</td><td>${config.half_day_salary}</td><td>${formatAmount(halfDaySalary)} kr</td></tr>`;
    }
    if (overnight > 0) {
      html += `<tr><td>Övernattning</td><td colspan="2">${overnight}</td><td>${config.overnight_salary}</td><td>${formatAmount(overnightSalary)} kr</td></tr>`;
    }
    if (salaryOvrigTid.hours > 0 || salaryOvrigTid.minutes > 0) {
      html += `<tr><td>Övrig tid</td><td>${salaryOvrigTid.hours}</td><td>${salaryOvrigTid.minutes}</td><td>${salaryOvrigTid.salary ?? '-'}</td><td>${formatAmount(salaryOvrigTid.total)} kr</td></tr>`;
    }
    // Fixed monetary addon (e.g. travel compensation) — added on top of the total
    const addonAmount = instructor.addon_amount && instructor.addon_description ? instructor.addon_amount : 0;
    if (addonAmount > 0) {
      html += `<tr><td>${escapeHtml(instructor.addon_description!)}</td><td colspan="3"></td><td>${formatAmount(addonAmount)} kr</td></tr>`;
    }
    html += `<tr style="font-weight:bold"><td>Totalt</td><td colspan="3"></td><td>${formatAmount(Math.round(totalSalary + fullDaySalary + halfDaySalary + overnightSalary + addonAmount))} kr</td></tr>`;
    html += `</tbody></table>`;
  }

  // Attachments
  if (attachments && attachments.length > 0) {
    let utlaggHtml = '';
    for (const att of attachments) {
      utlaggHtml += `<li>${escapeHtml(att.filename)}${att.desc ? ` – ${escapeHtml(att.desc)}` : ''}</li>`;
    }
    if (utlaggHtml) {
      html += `<h4>Utlägg</h4><ul>${utlaggHtml}</ul>`;
    }
  }

  return html;
}
