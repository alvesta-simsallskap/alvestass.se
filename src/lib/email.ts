// Email sending logic for Cloudflare/Edge (Mailjet API)
import type { TimeReportData } from './types';

export async function sendTimeReportEmail({
  data,
  attachments,
  MJ_APIKEY_PUBLIC,
  MJ_APIKEY_PRIVATE,
  html,
  monthKey
}: {
  data: TimeReportData;
  attachments: any[];
  MJ_APIKEY_PUBLIC: string;
  MJ_APIKEY_PRIVATE: string;
  html: string;
  monthKey: string;
}): Promise<Response> {
  const recipients = [{ Email: 'lon@alvestass.se' }];
  const ccRecipients = data.email ? [{ Email: data.email }] : [];
  const mailjetPayload = {
    Messages: [
      {
        From: { Email: 'noreply@alvestass.se', Name: 'Alvesta Simsällskap' },
        To: recipients,
        Cc: ccRecipients,
        Subject: `Tidrapport för ${data.name} ${monthKey}`,
        HTMLPart: html,
        Attachments: attachments.length > 0 ? attachments : undefined,
      },
    ],
  };
  return fetch('https://api.mailjet.com/v3.1/send', {
    method: 'POST',
    headers: {
      Authorization: 'Basic ' + btoa(`${MJ_APIKEY_PUBLIC}:${MJ_APIKEY_PRIVATE}`),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(mailjetPayload),
  });
}
