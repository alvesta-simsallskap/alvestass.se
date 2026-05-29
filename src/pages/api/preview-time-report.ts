export const prerender = false;

import type { APIRoute } from 'astro';
import { env } from 'cloudflare:workers';
import { parseTimeReportForm } from '../../lib/timeReportValidation';
import { buildTimeReportHtml } from '../../lib/timeReportHtml';
import { authenticateServiceUser, fetchTimeReportConfig, fetchTimeReportSessions, fetchInstructor } from '../../lib/trailbase';
import type { TimeReportConfig, Instructor, TrainingGroupKey, SessionSchedule } from '../../lib/types';

export const POST: APIRoute = async ({ request }) => {
  const TRAILBASE_URL = env.TRAILBASE_URL;
  const TRAILBASE_SERVICE_EMAIL = env.TRAILBASE_SERVICE_EMAIL;
  const TRAILBASE_SERVICE_PASSWORD = env.TRAILBASE_SERVICE_PASSWORD;

  const formData = await request.formData();

  // Parse form (no Turnstile verification for preview)
  const data = parseTimeReportForm(formData);

  // Fetch config, sessions and instructor from Trailbase (same graceful degrade as send endpoint)
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
  const effectiveConfig: TimeReportConfig = config ?? {
    id: 0,
    active_month_key: '',
    active_month_display: activeMonthDisplay,
    extra_time_simskola: 30,
    extra_time_training: 15,
    half_day_salary: 500,
    full_day_salary: 1000,
    overnight_salary: 300,
  };

  // Collect attachment names/descriptions for preview (not actual base64 content)
  const attachments: { filename: string; desc?: string }[] = [];
  for (const [key, value] of formData.entries()) {
    if (typeof value === 'object' && value instanceof File && key.startsWith('utlagg_file_')) {
      const id = key.replace('utlagg_file_', '');
      const desc = formData.get(`utlagg_desc_${id}`) as string | null;
      if (value.size > 0) {
        attachments.push({
          filename: value.name,
          desc: desc || undefined,
        });
      }
    }
  }

  // Build the HTML (no timestamp; that's only for the actual send)
  const html = buildTimeReportHtml(data, schedule, effectiveConfig, instructor, activeMonthDisplay, attachments);

  return new Response(html, { status: 200, headers: { 'Content-Type': 'text/html; charset=utf-8' } });
};
