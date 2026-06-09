// Shared types for time report API and frontend

export interface ExtraTimeRow {
  date: string;
  h: string;
  m: string;
  desc: string;
}

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
  extratid?: ExtraTimeRow[];
}

export interface Instructor {
  id: number;
  email: string;
  name: string;
  swim_school_rate: number | null;
  coach_rate: number | null;
  travel_compensation: boolean;
  addon_amount: number | null;
  addon_description: string | null;
}

export type TrainingGroupKey =
  | 'simskola'
  | 'tavlingA'
  | 'tavlingB'
  | 'teknik'
  | 'masters'
  | 'vuxencrawl';

export interface Session {
  date: string;
  title: string;
  hours: number;
  minutes: number;
}

export type SessionSchedule = Record<TrainingGroupKey, Session[]>;

export interface TimeReportConfig {
  id: number;
  active_month_key: string;
  active_month_display: string;
  extra_time_simskola: number;
  extra_time_training: number;
  half_day_salary: number;
  full_day_salary: number;
  overnight_salary: number;
}
