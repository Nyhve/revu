export type HabitFrequency = "DAILY" | "WEEKLY";

export interface Habit {
  id: number;
  name: string;
  description: string | null;
  frequency: HabitFrequency;
  createdAt: string;
}

export interface HabitEntry {
  id: number;
  habitId: number;
  date: string;
  createdAt: string;
}

export interface CreateHabitRequest {
  name: string;
  description?: string;
  frequency: HabitFrequency;
}