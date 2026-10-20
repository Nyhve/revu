import { api } from "./client";
import type { UserPreferences } from "../types/user";

export const userApi = {
  getPreferences: () => api.get<UserPreferences>("/user/preferences"),
  updatePreferences: (data: UserPreferences) =>
    api.put<UserPreferences>("/user/preferences", data),
};