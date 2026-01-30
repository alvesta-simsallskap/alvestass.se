import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request }) => {
  const formData = await request.formData();

  // Extract fields
  const name = formData.get('Namn') || '';
  const email = formData.get('E-post') || '';
  const milersattning = formData.get('milersattning') || '';
  const kommentarer = formData.get('kommentarer') || '';

  // Collect checked dates for each section
  const simskola = formData.getAll('simskola_checked_dates[]');
  const tavling = formData.getAll('tävling_checked_dates[]');
  const teknik = formData.getAll('teknik_checked_dates[]');
  const masters = formData.getAll('masters_checked_dates[]');
  const vuxencrawl = formData.getAll('vuxencrawl_checked_dates[]');
  const ovrigt = formData.getAll('övrigt_checked_dates[]');

  // Compose email content
  const text = `
    Namn: ${name}
    E-post: ${email}
    Milersättning: ${milersattning}
    Kommentarer: ${kommentarer}

    Simskola: ${simskola.join(', ')}
    Tävlingsgrupp: ${tavling.join(', ')}
    Teknik: ${teknik.join(', ')}
    Masters: ${masters.join(', ')}
    Vuxencrawl: ${vuxencrawl.join(', ')}
    Övrigt: ${ovrigt.join(', ')}
  `;

  // Mailjet API credentials from environment variables
  const MJ_APIKEY_PUBLIC = import.meta.env.MJ_APIKEY_PUBLIC;
  const MJ_APIKEY_PRIVATE = import.meta.env.MJ_APIKEY_PRIVATE;

  // Recipients
  const recipients = [
    { Email: "johan.marand@alvestass.se" },
    { Email: "johan.marand@icloud.com" }
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
          From: { Email: "noreply@alvestass.se", Name: "Tidrapport" },
          To: recipients,
          Subject: "Ny tidrapport",
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