import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { clearToken, getToken, setToken } from "../api/client";
import { userApi } from "../api/userApi";
import type { UserPreferences } from "../types/user";

export const DEFAULT_PREFERENCES: UserPreferences = {
  goalsEnabled: true,
  tasksEnabled: true,
  habitsEnabled: true,
  fitnessEnabled: true,
  learningEnabled: true,
  financeEnabled: true,
};

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(getToken());
  const preferences = ref<UserPreferences | null>(null);

  const isAuthenticated = computed(() => token.value !== null);

  /** Preferences with defaults while they are still loading. */
  const features = computed<UserPreferences>(() => preferences.value ?? DEFAULT_PREFERENCES);

  async function loadPreferences(): Promise<void> {
    try {
      preferences.value = await userApi.getPreferences();
    } catch {
      if (isAuthenticated.value) preferences.value = { ...DEFAULT_PREFERENCES };
    }
  }

  function login(newToken: string): void {
    setToken(newToken);
    token.value = newToken;
    void loadPreferences();
  }

  function logout(): void {
    clearToken();
    token.value = null;
    preferences.value = null;
  }

  function setPreferences(prefs: UserPreferences): void {
    preferences.value = prefs;
  }

  return { token, preferences, features, isAuthenticated, loadPreferences, login, logout, setPreferences };
});
