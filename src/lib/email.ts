// Email sending logic for Cloudflare/Edge (Mailjet API)
import type { TimeReportData, Employee } from './types';

export async function sendTimeReportEmail({
  data,
  employee,
  attachments,
  MJ_APIKEY_PUBLIC,
  MJ_APIKEY_PRIVATE,
  html,
}: {
  data: TimeReportData;
  employee: Employee | undefined;
  attachments: any[];
  MJ_APIKEY_PUBLIC: string;
  MJ_APIKEY_PRIVATE: string;
  html: string;
}): Promise<Response> {
  const recipients = [{ Email: 'lon@alvestass.se' }];
  const ccRecipients = data.email ? [{ Email: data.email }] : [];
  const mailjetPayload = {
    Messages: [
      {
        From: { Email: 'noreply@alvestass.se', Name: 'Alvesta Simsällskap' },
        To: recipients,
        Cc: ccRecipients,
        Subject: `Tidrapport för ${data.name} 2026-02`,
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
