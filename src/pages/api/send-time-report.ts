export const prerender = false;

import type { APIRoute } from 'astro';
import { parseTimeReportForm } from '../../lib/timeReportValidation';
import { sendTimeReportEmail } from '../../lib/email';
import { buildTable, calcSalary, findTimeItem } from '../../lib/salary';
import type { TimeReportData, Employee } from '../../lib/types';

export const POST: APIRoute = async ({ request, locals }) => {
  const EMPLOYEES: Employee[] = [
    { email: 'johan.marand@icloud.com', swimSchoolRate: 115, coachRate: 145 }, // Testing as William
    { email: 'annamaria.tovar@hotmail.com', swimSchoolRate: 85, coachRate: null },
    { email: 'cornelia.axhed@outlook.com', swimSchoolRate: 95, coachRate: 135 },
    { email: 'izakaxelsson2009@gmail.com', swimSchoolRate: 95, coachRate: null },
    { email: 'widqvistlova@gmail.com', swimSchoolRate: 95, coachRate: null },
    { email: 'thelmaklintipsa@gmail.com', swimSchoolRate: 115, coachRate: 135 },
    { email: 'jonnaklintipsa11@gmail.com', swimSchoolRate: 65, coachRate: null },
    { email: 'tsinatweldemichael42@gmail.com', swimSchoolRate: 85, coachRate: null },
    { email: 'edvin.nilsson.11@edualvesta.se', swimSchoolRate: 65, coachRate: null },
    { email: 'shahed27kikar@gmail.com', swimSchoolRate: 75, coachRate: null },
    { email: 'reussfelix390@gmail.com', swimSchoolRate: 75, coachRate: null },
    { email: 'alibrahemali06@gmail.com', swimSchoolRate: 65, coachRate: null },
    { email: 'aina.mujadzic@outlook.com', swimSchoolRate: 85, coachRate: 105 },
    { email: 'emii0113lus@gmail.com', swimSchoolRate: 75, coachRate: null },
    { email: 'stella.gustavsson@icloud.com', swimSchoolRate: 85, coachRate: 105 },
    { email: 'erik.hakamsson@gmail.com', swimSchoolRate: 75, coachRate: null },
    { email: 'ra7838303@gmail.com', swimSchoolRate: 75, coachRate: null },
    { email: 'aliciablyth@hotmail.com', swimSchoolRate: 95, coachRate: null },
    { email: 'erona.h09@gmail.com', swimSchoolRate: 75, coachRate: null },
    { email: 'williamlarson1999@gmail.com', swimSchoolRate: 115, coachRate: 145 },
    { email: 'mujcicuna@gmail.com', swimSchoolRate: 95, coachRate: 125 },
    { email: 'toweeandersson@gmail.com', swimSchoolRate: 105, coachRate: 135 },
    { email: 'agnesannaandersson@gmail.com', swimSchoolRate: 95, coachRate: 115 },
  ]
  
  const MJ_APIKEY_PUBLIC = locals.runtime.env.MJ_APIKEY_PUBLIC;
  const MJ_APIKEY_PRIVATE = locals.runtime.env.MJ_APIKEY_PRIVATE;
  const TURNSTILE_SECRET_KEY = locals.runtime.env.TURNSTILE_SECRET_KEY;

  // Detect debug mode (localhost)
  const isDebug = (request.headers.get('host')?.startsWith('localhost') ?? false);
 
  const formData = await request.formData();
  // Turnstile verification (skip in debug)
  if (!isDebug) {
    const token = formData.get('cf-turnstile-response');
    const secretKey = TURNSTILE_SECRET_KEY;
    const verifyRes = await fetch('https://challenges.cloudflare.com/turnstile/v0/siteverify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: `secret=${secretKey}&response=${token}`,
    });
    const verifyData = await verifyRes.json() as { success: boolean };
    if (!verifyData.success) {
      return new Response('Turnstile verification failed', { status: 400 });
    }
  }

  // Parse and validate form
  const data: TimeReportData = parseTimeReportForm(formData);


  // Compose email content (HTML)
  let html = '<h4>Tidrapport januari 2026</h4>';
  html += `<p><b>Namn:</b> ${data.name}</p>`;
  html += buildTable('simskola', 'Simskola', data.simskola);
  html += buildTable('tavlingA', 'Tävlingsgrupp A', data.tavlingA);
  html += buildTable('tavlingB', 'Tävlingsgrupp B', data.tavlingB);
  html += buildTable('teknik', 'Teknik', data.teknik);
  html += buildTable('masters', 'Masters', data.masters);
  html += buildTable('vuxencrawl', 'Vuxencrawl', data.vuxencrawl);
  if (data.milersattning) {
    html += `<p><b>Milersättning:</b> ${data.milersattning} km</p>`;
  }
  if (data.kommentarer) {
    html += `<p><b>Kommentarer:</b> ${data.kommentarer}</p>`;
  }

  // Find employee salary info
  const employee = EMPLOYEES.find(e => e.email.toLowerCase() === String(data.email).toLowerCase());


  // Calculate salary for each section
  const salarySimskola = calcSalary('simskola', data.simskola, employee);
  const salaryTavlingA = calcSalary('tavlingA', data.tavlingA, employee);
  const salaryTavlingB = calcSalary('tavlingB', data.tavlingB, employee);
  const salaryTeknik = calcSalary('teknik', data.teknik, employee);
  const salaryMasters = calcSalary('masters', data.masters, employee);
  const salaryVuxencrawl = calcSalary('vuxencrawl', data.vuxencrawl, employee);

  // Calculate total salary
  const totalSalary = [salarySimskola, salaryTavlingA, salaryTavlingB, salaryTeknik, salaryMasters, salaryVuxencrawl]
    .reduce((sum, s) => sum + s.total, 0);

  // Calculate full day and half day
  let fullDay = 0, halfDay = 0;
  for (const val of data.tavlingA) {
    const item = findTimeItem('tavlingA', val);
    if (item?.h === 20) { fullDay++; }
    if (item?.h === 10) { halfDay++; }
  }
  for (const val of data.tavlingB) {
    const item = findTimeItem('tavlingB', val);
    if (item?.h === 20) { fullDay++; }
    if (item?.h === 10) { halfDay++; }
  }
  for (const val of data.masters) {
    const item = findTimeItem('masters', val);
    if (item?.h === 20) { fullDay++; }
    if (item?.h === 10) { halfDay++; }
  }
  for (const val of data.teknik) {
    const item = findTimeItem('teknik', val);
    if (item?.h === 20) { fullDay++; }
    if (item?.h === 10) { halfDay++; }
  }

  const fullDaySalary = 1000 * fullDay;
  const halfDaySalary = 500 * halfDay;

  // Add salary estimate to email content if employee matched
  // Helper to format numbers with space as thousands separator
  const formatAmount = (amount: number) => amount.toLocaleString('sv-SE');

  if (employee) {
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
      html += `<tr><td>Heldagar</td><td colspan="2">${fullDay}</td><td>1 000</td><td>${formatAmount(fullDaySalary)} kr</td></tr>`;
    }
    if (halfDay > 0) {
      html += `<tr><td>Halvdagar</td><td colspan="2">${halfDay}</td><td>500</td><td>${formatAmount(halfDaySalary)} kr</td></tr>`;
    }
    html += `<tr style="font-weight:bold"><td>Totalt</td><td colspan="3"></td><td>${formatAmount(Math.round(totalSalary + fullDaySalary + halfDaySalary))} kr</td></tr>`;
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
  if (isDebug) {
    return new Response(html, { status: 200, headers: { 'Content-Type': 'text/html; charset=utf-8' } });
  }

  // Send email via utility
  const res = await sendTimeReportEmail({
    data,
    employee,
    attachments,
    MJ_APIKEY_PUBLIC,
    MJ_APIKEY_PRIVATE,
    html,
  });

  if (res.ok) {
    return new Response('OK', { status: 200 });
  } else {
    const error = await res.text();
    return new Response('Failed to send email: ' + error, { status: 500 });
  }
};