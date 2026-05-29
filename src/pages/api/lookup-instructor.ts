export const prerender = false;

import type { APIRoute } from 'astro';
import { env } from 'cloudflare:workers';
import { authenticateServiceUser, fetchInstructor } from '../../lib/trailbase';

export const POST: APIRoute = async ({ request }) => {
  const headers = { 'Cache-Control': 'no-store', 'Content-Type': 'application/json' };

  let email: string;
  try {
    const body: unknown = await request.json();
    if (
      typeof body !== 'object' ||
      body === null ||
      typeof (body as Record<string, unknown>).email !== 'string' ||
      !(body as Record<string, string>).email.includes('@')
    ) {
      return new Response(JSON.stringify({ error: 'invalid_email' }), { status: 400, headers });
    }
    email = (body as Record<string, string>).email.toLowerCase().trim();
  } catch {
    return new Response(JSON.stringify({ error: 'invalid_email' }), { status: 400, headers });
  }

  let instructor;
  try {
    const authToken = await authenticateServiceUser(
      env.TRAILBASE_URL,
      env.TRAILBASE_SERVICE_EMAIL,
      env.TRAILBASE_SERVICE_PASSWORD,
    );
    instructor = await fetchInstructor(env.TRAILBASE_URL, email, authToken);
  } catch {
    return new Response(JSON.stringify({ error: 'backend_unavailable' }), { status: 503, headers });
  }

  if (!instructor) {
    return new Response(JSON.stringify({ error: 'not_found' }), { status: 404, headers });
  }

  return new Response(
    JSON.stringify({
      swimSchool: instructor.swim_school_rate !== null,
      coach: instructor.coach_rate !== null,
      travelCompensation: Boolean(instructor.travel_compensation),
    }),
    { status: 200, headers },
  );
};
