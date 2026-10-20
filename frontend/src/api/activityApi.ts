import { api } from "./client";
import type { ActivityResponse } from "../types/activity";

export interface DailyActivityCount {
  date: string;
  count: number;
}

export const activityApi = {
  getFeed: () => api.get<ActivityResponse[]>("/activities"),
  getDailyCounts: (days: number) =>
    api.get<DailyActivityCount[]>(`/activities/daily?days=${days}`),
};