<script setup lang="ts">
import { onMounted, ref } from "vue";
import { userApi } from "../api/userApi";
import { errorMessage } from "../api/client";
import type { UserPreferences } from "../types/user";
import { useAuthStore } from "../stores/auth";
import { useThemeStore } from "../stores/theme";
import ToggleSwitch from "../components/ToggleSwitch.vue";

const FEATURES: { key: keyof UserPreferences; label: string; description: string }[] = [
  { key: "goalsEnabled", label: "Ziele", description: "Langfristige Lebensziele verfolgen" },
  { key: "tasksEnabled", label: "Aufgaben", description: "To-do Liste für kurzfristige Aufgaben" },
  { key: "habitsEnabled", label: "Habits", description: "Tägliche und wöchentliche Gewohnheiten" },
  { key: "fitnessEnabled", label: "Fitness", description: "Workouts und Trainingsfortschritt" },
  { key: "learningEnabled", label: "Learning", description: "Lernprojekte und Lernzeit tracken" },
  { key: "financeEnabled", label: "Finanzen", description: "Einnahmen und Ausgaben verwalten" },
];

const auth = useAuthStore();
const theme = useThemeStore();

const preferences = ref<UserPreferences | null>(null);
const isLoading = ref(true);
const isSaving = ref(false);
const error = ref<string | null>(null);
const saved = ref(false);
let savedTimer: ReturnType<typeof setTimeout> | undefined;

onMounted(async () => {
  try {
    preferences.value = await userApi.getPreferences();
  } catch (err) {
    error.value = errorMessage(err, "Einstellungen konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
});

async function toggleFeature(key: keyof UserPreferences) {
  if (!preferences.value) return;
  const previous = preferences.value;
  const updated = { ...previous, [key]: !previous[key] };
  preferences.value = updated;
  isSaving.value = true;
  saved.value = false;
  error.value = null;
  try {
    const result = await userApi.updatePreferences(updated);
    preferences.value = result;
    auth.setPreferences(result);
    saved.value = true;
    clearTimeout(savedTimer);
    savedTimer = setTimeout(() => (saved.value = false), 2000);
  } catch (err) {
    error.value = errorMessage(err, "Einstellungen konnten nicht gespeichert werden.");
    preferences.value = previous;
  } finally {
    isSaving.value = false;
  }
}
</script>

<template>
  <div class="max-w-lg">
    <h1 class="text-2xl font-semibold mb-8">Profil</h1>

    <p v-if="error" class="text-sm text-red-700 mb-4">{{ error }}</p>

    <div class="space-y-8">
      <section>
        <h2 class="text-sm font-medium text-ink-muted mb-4">Erscheinungsbild</h2>
        <div class="flex items-center justify-between py-3 border-b border-border">
          <div>
            <p class="text-sm font-medium">Dark Mode</p>
            <p class="text-xs text-ink-muted mt-0.5">Dunkles Farbschema verwenden</p>
          </div>
          <ToggleSwitch :model-value="theme.isDark" label="Dark Mode umschalten" @update:model-value="theme.toggle()" />
        </div>
      </section>

      <p v-if="isLoading" class="text-ink-muted">Lädt…</p>
      <section v-else-if="preferences">
        <h2 class="text-sm font-medium text-ink-muted mb-1">Features</h2>
        <p class="text-xs text-ink-muted mb-4">
          Deaktivierte Features werden aus der Navigation und der Übersicht ausgeblendet.
        </p>
        <div class="divide-y divide-border border-t border-b border-border">
          <div v-for="f in FEATURES" :key="f.key" class="flex items-center justify-between py-4">
            <div>
              <p class="text-sm font-medium" :class="{ 'text-ink-muted': !preferences[f.key] }">{{ f.label }}</p>
              <p class="text-xs text-ink-muted mt-0.5">{{ f.description }}</p>
            </div>
            <ToggleSwitch
              :model-value="preferences[f.key]"
              :disabled="isSaving"
              :label="`${f.label} ${preferences[f.key] ? 'deaktivieren' : 'aktivieren'}`"
              @update:model-value="toggleFeature(f.key)"
            />
          </div>
        </div>
        <p v-if="saved" class="text-xs text-accent mt-3">✓ Gespeichert</p>
      </section>
    </div>
  </div>
</template>
