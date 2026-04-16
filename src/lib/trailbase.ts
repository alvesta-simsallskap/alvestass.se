export interface ClubInfo {
  id: number;
  name: string;
  tagline: string;
  founding_year: number;
  short_description: string;
  address: string;
  city: string;
  postal_code: string;
  phone: string;
  email: string;
}

interface TrailbaseListResponse {
  records: ClubInfo[];
  cursor: string | null;
}

/**
 * Fetch the single club_info record from Trailbase.
 * Returns `null` if the table is empty (no record yet).
 * Throws on network errors — the caller decides on fallback.
 */
export async function fetchClubInfo(baseUrl: string): Promise<ClubInfo | null> {
  const response = await fetch(
    `${baseUrl}/api/records/v1/club_info?limit=1`
  );

  if (!response.ok) {
    throw new Error(`Trailbase responded with ${response.status}`);
  }

  const body: TrailbaseListResponse = await response.json();

  if (!body.records || body.records.length === 0) {
    return null;
  }

  return body.records[0];
}
