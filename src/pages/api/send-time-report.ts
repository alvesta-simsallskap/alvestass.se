export const prerender = false;
import timeReportItems from '../../config/time-report-items.json';
import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request, locals }) => {
  const EMPLOYEES = [
    { email: 'johan.marand@icloud.com', swimSchoolRate: 85, coachRate: 105 },
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
    { email: 'williamlarson1999@gmail.com', swimSchoolRate: 115, coachRate: 155 },
    { email: 'mujcicuna@gmail.com', swimSchoolRate: 95, coachRate: 125 },
    { email: 'toweeandersson@gmail.com', swimSchoolRate: 105, coachRate: 135 },
    { email: 'agnesannaandersson@gmail.com', swimSchoolRate: 95, coachRate: 115 },
  ]
  
  const MJ_APIKEY_PUBLIC = locals.runtime.env.MJ_APIKEY_PUBLIC;
  const MJ_APIKEY_PRIVATE = locals.runtime.env.MJ_APIKEY_PRIVATE;
  const TURNSTILE_SECRET_KEY = locals.runtime.env.TURNSTILE_SECRET_KEY;
 
  const formData = await request.formData();

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

  // Extract fields
  const name = formData.get('namn') || '';
  const email = formData.get('email') || '';
  const milersattning = formData.get('milersattning') || '';
  const kommentarer = formData.get('kommentarer') || '';

  // Collect checked dates for each section (supporting new keys)
  const simskola = formData.getAll('simskola_checked_dates[]').filter((v): v is string => typeof v === 'string');
  const tavlingA = formData.getAll('tavling-a_checked_dates[]').filter((v): v is string => typeof v === 'string');
  const tavlingB = formData.getAll('tavling-b_checked_dates[]').filter((v): v is string => typeof v === 'string');
  const teknik = formData.getAll('teknik_checked_dates[]').filter((v): v is string => typeof v === 'string');
  const masters = formData.getAll('masters_checked_dates[]').filter((v): v is string => typeof v === 'string');
  const vuxencrawl = formData.getAll('vuxencrawl_checked_dates[]').filter((v): v is string => typeof v === 'string');

  // Helper to find time info for a checked item
  function findTimeItem(section: string, value: string) {
    // value is "YYYY-MM-DD Title"
    const [date, ...titleParts] = value.split(' ');
    const title = titleParts.join(' ');
    const items = (timeReportItems['2026-01'] as any)[section] || [];
    return items.find((item: any) => item.date === date && item.title === title);
  }

  // Helper to build HTML table for a section
  function buildTable(section: string, label: string, checked: string[]) {
    if (!checked.length) return '';
    let rows = '';
    for (const val of checked) {
      const item = findTimeItem(section, val);
      if (item) {
        let time;
        if (item.h === 20) {
          time = 'Heldag';
        } else if (item.h === 10) {
          time = 'Halvdag';
        } else {
          time = `${item.h}:${item.m < 10 ? '0' : ''}${item.m}`;
        }
        
        rows += `<tr><td>${val}</td><td>${time}</td></tr>`;
      } else {
        rows += `<tr><td>${val}</td><td></td></tr>`;
      }
    }
    return `<h4>${label}</h4><table border="1" cellpadding="4" style="border-collapse:collapse;margin-bottom:1em;"><thead><tr><th>Datum och aktivitet</th><th>Tid</th></tr></thead><tbody>${rows}</tbody></table>`;
  }

  // Compose email content (HTML)
  let html = '<h4>Tidrapport januari 2026</h4>'
  html += `<p><b>Namn:</b> ${name}</p>`;
  html += buildTable('simskola', 'Simskola', simskola);
  html += buildTable('tavlingA', 'Tävlingsgrupp A', tavlingA);
  html += buildTable('tavlingB', 'Tävlingsgrupp B', tavlingB);
  html += buildTable('teknik', 'Teknik', teknik);
  html += buildTable('masters', 'Masters', masters);
  html += buildTable('vuxencrawl', 'Vuxencrawl', vuxencrawl);
  if (milersattning) {
    html += `<p><b>Milersättning:</b> ${milersattning} km</p>`;
  }
  if (kommentarer) {
    html += `<p><b>Kommentarer:</b> ${kommentarer}</p>`;
  }

  // Find employee salary info
  const employee = EMPLOYEES.find(e => e.email.toLowerCase() === String(email).toLowerCase());

  // Helper to calculate salary for a section
  function calcSalary(section: string, checked: string[]): { hours: number, minutes: number, salary: number|null, total: number } {
    let hours = 0, minutes = 0;
    let rate: number|null = null;
    if (!employee) return { hours, minutes, salary: null, total: 0 };
    if (section === 'simskola') {
      rate = employee.swimSchoolRate;
    } else {
      rate = employee.coachRate;
    }
    for (const val of checked) {
      const item = findTimeItem(section, val);
      if (item && typeof item.h === 'number' && typeof item.m === 'number') {
        hours += item.h;
        minutes += item.m;
      }
    }
    const total = rate ? (hours + minutes / 60) * rate : 0;
    return { hours, minutes, salary: rate, total };
  }

  // Calculate salary for each section
  const salarySimskola = calcSalary('simskola', simskola);
  const salaryTavlingA = calcSalary('tavlingA', tavlingA);
  const salaryTavlingB = calcSalary('tavlingB', tavlingB);
  const salaryTeknik = calcSalary('teknik', teknik);
  const salaryMasters = calcSalary('masters', masters);
  const salaryVuxencrawl = calcSalary('vuxencrawl', vuxencrawl);

  // Calculate total salary
  const totalSalary = [salarySimskola, salaryTavlingA, salaryTavlingB, salaryTeknik, salaryMasters, salaryVuxencrawl]
    .reduce((sum, s) => sum + s.total, 0);

  // Add salary estimate to email content if employee matched
  if (employee) {
    html += `<h4>Preliminär löneberäkning</h4><table border="1" cellpadding="4" style="border-collapse:collapse;margin-bottom:1em;">
      <thead><tr><th>Grupp</th><th>Timmar</th><th>Minuter</th><th>Timlön</th><th>Summa</th></tr></thead><tbody>
      <tr><td>Simskola</td><td>${salarySimskola.hours}</td><td>${salarySimskola.minutes}</td><td>${salarySimskola.salary ?? '-'}</td><td>${salarySimskola.total.toFixed(2)} kr</td></tr>
      <tr><td>Tävlingsgrupp A</td><td>${salaryTavlingA.hours}</td><td>${salaryTavlingA.minutes}</td><td>${salaryTavlingA.salary ?? '-'}</td><td>${salaryTavlingA.total.toFixed(2)} kr</td></tr>
      <tr><td>Tävlingsgrupp B</td><td>${salaryTavlingB.hours}</td><td>${salaryTavlingB.minutes}</td><td>${salaryTavlingB.salary ?? '-'}</td><td>${salaryTavlingB.total.toFixed(2)} kr</td></tr>
      <tr><td>Teknik</td><td>${salaryTeknik.hours}</td><td>${salaryTeknik.minutes}</td><td>${salaryTeknik.salary ?? '-'}</td><td>${salaryTeknik.total.toFixed(2)} kr</td></tr>
      <tr><td>Masters</td><td>${salaryMasters.hours}</td><td>${salaryMasters.minutes}</td><td>${salaryMasters.salary ?? '-'}</td><td>${salaryMasters.total.toFixed(2)} kr</td></tr>
      <tr><td>Vuxencrawl</td><td>${salaryVuxencrawl.hours}</td><td>${salaryVuxencrawl.minutes}</td><td>${salaryVuxencrawl.salary ?? '-'}</td><td>${salaryVuxencrawl.total.toFixed(2)} kr</td></tr>
      <tr style="font-weight:bold"><td>Totalt</td><td colspan="3"></td><td>${Math.round(totalSalary)} kr</td></tr>
      </tbody></table>`;
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
  html += `<p><i>Tidrapporten skickades in genom alvestass.se/tidrapport ${formattedDate}</i></p>`;

  // Recipients
  const recipients = [
    { Email: "lon@alvestass.se" }
  ];
  const ccRecipients = email ? [{ Email: email }] : [];

  // Send email via Mailjet API (HTML + attachments)
  const mailjetPayload = {
    Messages: [
      {
        From: { Email: "noreply@alvestass.se", Name: "Alvesta Simsällskap" },
        To: recipients,
        Cc: ccRecipients,
        Subject: `Tidrapport för ${name} 2026-01`,
        HTMLPart: html,
        Attachments: attachments.length > 0 ? attachments : undefined
      }
    ]
  };

  const res = await fetch('https://api.mailjet.com/v3.1/send', {
    method: 'POST',
    headers: {
      'Authorization': 'Basic ' + btoa(`${MJ_APIKEY_PUBLIC}:${MJ_APIKEY_PRIVATE}`),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(mailjetPayload),
  });

  if (res.ok) {
    return new Response('OK', { status: 200 });
  } else {
    const error = await res.text();
    return new Response('Failed to send email: ' + error, { status: 500 });
  }
};