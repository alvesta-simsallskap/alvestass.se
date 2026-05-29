export const prerender = false;

import type { APIRoute } from 'astro';
import { env } from 'cloudflare:workers';
import { parseTimeReportForm } from '../../lib/timeReportValidation';
import { sendTimeReportEmail } from '../../lib/email';
import { buildTimeReportHtml } from '../../lib/timeReportHtml';
import { authenticateServiceUser, fetchTimeReportConfig, fetchTimeReportSessions, fetchInstructor } from '../../lib/trailbase';
import type { TimeReportData, Instructor, TrainingGroupKey, SessionSchedule, TimeReportConfig } from '../../lib/types';

export const POST: APIRoute = async ({ request }) => {
  const MJ_APIKEY_PUBLIC = env.MJ_APIKEY_PUBLIC;
  const MJ_APIKEY_PRIVATE = env.MJ_APIKEY_PRIVATE;
  const TURNSTILE_SECRET_KEY = env.TURNSTILE_SECRET_KEY;
  const TRAILBASE_URL = env.TRAILBASE_URL;
  const TRAILBASE_SERVICE_EMAIL = env.TRAILBASE_SERVICE_EMAIL;
  const TRAILBASE_SERVICE_PASSWORD = env.TRAILBASE_SERVICE_PASSWORD;

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

  // Collect file attachments for Mailjet (base64 encoded)
  const mailjetAttachments = [];
  for (const [key, value] of formData.entries()) {
    if (typeof value === 'object' && value instanceof File && key.startsWith('utlagg_file_')) {
      if (value.size > 0) {
        const arrayBuffer = await value.arrayBuffer();
        const base64 = Buffer.from(arrayBuffer).toString('base64');
        mailjetAttachments.push({
          ContentType: value.type,
          Filename: value.name,
          Base64Content: base64
        });
      }
    }
  }

  // Build the HTML report
  const html = buildTimeReportHtml(data, schedule, effectiveConfig, instructor, activeMonthDisplay);

  // Append timestamp (only for actual send, not preview)
  const formattedDate = new Date().toLocaleString('sv-SE', {
    timeZone: 'Europe/Stockholm',
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  }).replace(' ', ' kl. ').replace(':', '.');
  const htmlWithTimestamp = html + `<p><i>Skickades genom alvestass.se/tidrapport ${formattedDate}</i></p>`;

  // Send email via utility
  const res = await sendTimeReportEmail({
    data,
    attachments: mailjetAttachments,
    MJ_APIKEY_PUBLIC,
    MJ_APIKEY_PRIVATE,
    html: htmlWithTimestamp,
    monthKey: activeMonthKey,
  });

  if (res.ok) {
    return new Response('OK', { status: 200 });
  } else {
    const error = await res.text();
    return new Response('Failed to send email: ' + error, { status: 500 });
  }
};
