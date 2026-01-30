export const prerender = false;

import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request, locals }) => {
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

  // Collect checked dates for each section
  const simskola = formData.getAll('simskola_checked_dates[]');
  const tavling = formData.getAll('tavling_checked_dates[]');
  const teknik = formData.getAll('teknik_checked_dates[]');
  const masters = formData.getAll('masters_checked_dates[]');
  const vuxencrawl = formData.getAll('vuxencrawl_checked_dates[]');
  const ovrigt = formData.getAll('ovrigt_checked_dates[]');

  // Compose email content
  let text = `Namn: ${name}.\n E-post: ${email}\n\n`;

  if (simskola.length > 0) {
    text += `Simskola:\n${simskola.map(date => `  ${date}`).join('\n')}\n\n`;
  }
  if (tavling.length > 0) {
    text += `Tävlingsgrupp:\n${tavling.map(date => `  ${date}`).join('\n')}\n\n`;
  }
  if (teknik.length > 0) {
    text += `Teknik:\n${teknik.map(date => `  ${date}`).join('\n')}\n\n`;
  }
  if (masters.length > 0) {
    text += `Masters:\n${masters.map(date => `  ${date}`).join('\n')}\n\n`;
  }
  if (vuxencrawl.length > 0) {
    text += `Vuxencrawl:\n${vuxencrawl.map(date => `  ${date}`).join('\n')}\n\n`;
  }
  if (ovrigt.length > 0) {
    text += `Övrigt:\n${ovrigt.map(date => `  ${date}`).join('\n')}\n\n`;
  }

  if (milersattning) {
    text += `\nMilersättning: ${milersattning} km\n\n`;
  }

  if (kommentarer) {
    text += `\nKommentarer: ${kommentarer}\n`;
  }

  // Recipients
  const recipients = [
    { Email: "lon@alvestass.se" },
    { Email: "johan.marand@alvestass.se" },
    { Email: email } // The one who filled out the form
  ];

  // Send email via Mailjet API
  const res = await fetch('https://api.mailjet.com/v3.1/send', {
    method: 'POST',
    headers: {
      'Authorization': 'Basic ' + btoa(`${MJ_APIKEY_PUBLIC}:${MJ_APIKEY_PRIVATE}`),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      Messages: [
        {
          From: { Email: "noreply@alvestass.se", Name: "Alvesta Simsällskap" },
          To: recipients,
          Subject: `Tidrapport för ${name}`,
          TextPart: text,
        }
      ]
    }),
  });

  if (res.ok) {
    return new Response('OK', { status: 200 });
  } else {
    const error = await res.text();
    return new Response('Failed to send email: ' + error, { status: 500 });
  }
};