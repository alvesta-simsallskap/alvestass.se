// Shared types for time report API and frontend

export interface TimeReportData {
  name: string;
  email: string;
  milersattning: string;
  kommentarer: string;
  simskola: string[];
  tavlingA: string[];
  tavlingB: string[];
  teknik: string[];
  masters: string[];
  vuxencrawl: string[];
}

export interface Employee {
  email: string;
  swimSchoolRate: number;
  coachRate: number | null;
}
