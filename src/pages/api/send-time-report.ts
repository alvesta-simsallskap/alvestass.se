export const prerender = false;

import type { APIRoute } from 'astro';
import { parseTimeReportForm } from '../../lib/timeReportValidation';
import { sendTimeReportEmail } from '../../lib/email';
import { buildTable, buildAdditionalTimeTable, calcSalary, findTimeItem } from '../../lib/salary';
import { authenticateServiceUser, fetchTimeReportConfig, fetchTimeReportSessions, fetchInstructor } from '../../lib/trailbase';
import type { TimeReportData, Instructor, TrainingGroupKey, SessionSchedule, TimeReportConfig } from '../../lib/types';

export const POST: APIRoute = async ({ request, locals }) => {
  const MJ_APIKEY_PUBLIC = locals.runtime.env.MJ_APIKEY_PUBLIC;
  const MJ_APIKEY_PRIVATE = locals.runtime.env.MJ_APIKEY_PRIVATE;
  const TURNSTILE_SECRET_KEY = locals.runtime.env.TURNSTILE_SECRET_KEY;
  const TRAILBASE_URL = locals.runtime.env.TRAILBASE_URL;
  const TRAILBASE_SERVICE_EMAIL = locals.runtime.env.TRAILBASE_SERVICE_EMAIL;
  const TRAILBASE_SERVICE_PASSWORD = locals.runtime.env.TRAILBASE_SERVICE_PASSWORD;

  const formData = await request.formData();

  // Turnstile verification (skip in development mode)
  if (!import.meta.env.DEV) {
    const token = formData.get('cf-turnstile-response');
    const verifyRes = await fetch('https://challenges.cloudflare.com/turnstile/v0/siteverify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: `secret=${TURNSTILE_SECRET_KEY}&response=${token}`,
    });
    const verifyData = await verifyRes.json() as { success: boolean };
    if (!verifyData.success) {
      return new Response('Turnstile verification failed', { status: 400 });
    }
  }

  // Parse and validate form
  const data: TimeReportData = parseTimeReportForm(formData);

  // Fetch config, sessions and instructor from Trailbase.
  // Auth failure or missing config degrades gracefully — email is still sent
  // without a salary estimate (spec: a missing estimate is recoverable, a lost
  // submission is not).
  let instructor: Instructor | undefined;
  let config: TimeReportConfig | null = null;
  let schedule: SessionSchedule = {
    simskola: [], tavlingA: [], tavlingB: [], teknik: [], masters: [], vuxencrawl: [],
  };

  try {
    const authToken = await authenticateServiceUser(TRAILBASE_URL, TRAILBASE_SERVICE_EMAIL, TRAILBASE_SERVICE_PASSWORD);
    config = await fetchTimeReportConfig(TRAILBASE_URL, authToken);

    if (config) {
      const sessions = await fetchTimeReportSessions(TRAILBASE_URL, config.active_month_key, authToken);
      for (const session of sessions) {
        const group = session.training_group as TrainingGroupKey;
        if (group in schedule) {
          schedule[group].push({ date: session.date, title: session.title, hours: session.hours, minutes: session.minutes });
        }
      }
    }

    const found = await fetchInstructor(TRAILBASE_URL, data.email, authToken);
    if (found) instructor = found;
  } catch {
    // Backend unreachable — proceed without salary estimate
  }

  // Fall back to empty config values if Trailbase was unreachable
  const activeMonthDisplay = config?.active_month_display ?? '';
  const activeMonthKey = config?.active_month_key ?? '';
  const effectiveConfig: TimeReportConfig = config ?? {
    id: 0,
    active_month_key: activeMonthKey,
    active_month_display: activeMonthDisplay,
    extra_time_simskola: 30,
    extra_time_training: 15,
    half_day_salary: 500,
    full_day_salary: 1000,
    overnight_salary: 300,
  };

  // Compose email content (HTML)
  let html = `<h4>Tidrapport ${activeMonthDisplay}</h4>`;
  html += `<p><b>Namn:</b> ${data.name}</p>`;
  html += buildTable('simskola', 'Simskola', data.simskola, schedule, effectiveConfig);
  html += buildTable('tavlingA', 'Tävlingsgrupp A', data.tavlingA, schedule, effectiveConfig);
  html += buildTable('tavlingB', 'Tävlingsgrupp B', data.tavlingB, schedule, effectiveConfig);
  html += buildTable('teknik', 'Teknik', data.teknik, schedule, effectiveConfig);
  html += buildTable('masters', 'Masters', data.masters, schedule, effectiveConfig);
  html += buildTable('vuxencrawl', 'Vuxencrawl', data.vuxencrawl, schedule, effectiveConfig);

  if (data.extratid) {
    html += buildAdditionalTimeTable(data.extratid);
  }

  if (data.milersattning) {
    html += `<p><b>Milersättning:</b> ${data.milersattning} km (privat resa, skattepliktigt, 25 kr/mil)</p>`;
  }

  if (data.kommentarer) {
    html += `<p><b>Kommentarer:</b> ${data.kommentarer}</p>`;
  }

  // Calculate salary for each group
  const salarySimskola   = calcSalary('simskola',   data.simskola,   schedule, effectiveConfig, instructor);
  const salaryTavlingA   = calcSalary('tavlingA',   data.tavlingA,   schedule, effectiveConfig, instructor);
  const salaryTavlingB   = calcSalary('tavlingB',   data.tavlingB,   schedule, effectiveConfig, instructor);
  const salaryTeknik     = calcSalary('teknik',     data.teknik,     schedule, effectiveConfig, instructor);
  const salaryMasters    = calcSalary('masters',    data.masters,    schedule, effectiveConfig, instructor);
  const salaryVuxencrawl = calcSalary('vuxencrawl', data.vuxencrawl, schedule, effectiveConfig, instructor);
  const salaryOvrigTid   = calcSalary('extratid',   [],              schedule, effectiveConfig, instructor, data.extratid);

  const totalSalary = [salarySimskola, salaryTavlingA, salaryTavlingB, salaryTeknik, salaryMasters, salaryVuxencrawl, salaryOvrigTid]
    .reduce((sum, s) => sum + s.total, 0);

  // Count flat-rate competition sessions
  let fullDay = 0, halfDay = 0, overnight = 0;
  const countSpecial = (group: TrainingGroupKey, vals: string[]) => {
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

  const fullDaySalary   = effectiveConfig.full_day_salary * fullDay;
  const halfDaySalary   = effectiveConfig.half_day_salary * halfDay;
  const overnightSalary = effectiveConfig.overnight_salary * overnight;

  // Add salary estimate if instructor was found
  const formatAmount = (amount: number) => amount.toLocaleString('sv-SE', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

  if (instructor) {
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
      html += `<tr><td>Heldagar</td><td colspan="2">${fullDay}</td><td>${effectiveConfig.full_day_salary}</td><td>${formatAmount(fullDaySalary)} kr</td></tr>`;
    }
    if (halfDay > 0) {
      html += `<tr><td>Halvdagar</td><td colspan="2">${halfDay}</td><td>${effectiveConfig.half_day_salary}</td><td>${formatAmount(halfDaySalary)} kr</td></tr>`;
    }
    if (overnight > 0) {
      html += `<tr><td>Övernattning</td><td colspan="2">${overnight}</td><td>${effectiveConfig.overnight_salary}</td><td>${formatAmount(overnightSalary)} kr</td></tr>`;
    }
    if (salaryOvrigTid.hours > 0 || salaryOvrigTid.minutes > 0) {
      html += `<tr><td>Övrig tid</td><td>${salaryOvrigTid.hours}</td><td>${salaryOvrigTid.minutes}</td><td>${salaryOvrigTid.salary ?? '-'}</td><td>${formatAmount(salaryOvrigTid.total)} kr</td></tr>`;
    }
    html += `<tr style="font-weight:bold"><td>Totalt</td><td colspan="3"></td><td>${formatAmount(Math.round(totalSalary + fullDaySalary + halfDaySalary + overnightSalary))} kr</td></tr>`;
    html += `</tbody></table>`;
  }

  // Handle file attachments for 'Utlägg'
  const attachments = [];
  let utlaggHtml = '';
  for (const [key, value] of formData.entries()) {
    if (typeof value === 'object' && value instanceof File && key.startsWith('utlagg_file_')) {
      const id = key.replace('utlagg_file_', '');
      const desc = formData.get(`utlagg_desc_${id}`) || '';
      if (value.size > 0) {
        const arrayBuffer = await value.arrayBuffer();
        const base64 = Buffer.from(arrayBuffer).toString('base64');
        attachments.push({
          ContentType: value.type,
          Filename: value.name,
          Base64Content: base64
        });
        utlaggHtml += `<li>${value.name}${desc ? ` – ${desc}` : ''}</li>`;
      }
    }
  }
  if (utlaggHtml) {
    html += `<h4>Utlägg</h4><ul>${utlaggHtml}</ul>`;
  }

  // Information about the sending of the report
  const formattedDate = new Date().toLocaleString('sv-SE', {
    timeZone: 'Europe/Stockholm',
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  }).replace(' ', ' kl. ').replace(':', '.');
  html += `<p><i>Skickades genom alvestass.se/tidrapport ${formattedDate}</i></p>`;

  // In debug mode, show HTML output instead of sending email
  if (import.meta.env.DEV) {
    return new Response(html, { status: 200, headers: { 'Content-Type': 'text/html; charset=utf-8' } });
  }

  // Send email via utility
  const res = await sendTimeReportEmail({
    data,
    attachments,
    MJ_APIKEY_PUBLIC,
    MJ_APIKEY_PRIVATE,
    html,
    monthKey: activeMonthKey,
  });

  if (res.ok) {
    return new Response('OK', { status: 200 });
  } else {
    const error = await res.text();
    return new Response('Failed to send email: ' + error, { status: 500 });
  }
};
