import { api } from "./client";
import type { CreateHabitRequest, Habit, HabitEntry } from "../types/habit";

export const habitApi = {
  getAll: () => api.get<Habit[]>("/habits"),
  getTodaysEntries: () => api.get<HabitEntry[]>("/habits/entries/today"),
  create: (data: CreateHabitRequest) => api.post<Habit>("/habits", data),
  delete: (id: number) => api.delete<void>(`/habits/${id}`),
  completeToday: (id: number) => api.post<HabitEntry>(`/habits/${id}/entries`),
  getEntries: (id: number) => api.get<HabitEntry[]>(`/habits/${id}/entries`),
  deleteEntry: (habitId: number, entryId: number) =>
    api.delete<void>(`/habits/${habitId}/entries/${entryId}`),
  getConsistency: (days: number) =>
  api.get<HabitConsistency[]>(`/habits/consistency?days=${days}`),
};

export interface HabitConsistency {
  habitId: number;
  name: string;
  frequency: string;
  totalDays: number;
  completedDays: number;
  consistencyPercent: number;
}