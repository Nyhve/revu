<script setup lang="ts">
import { onMounted, ref } from "vue";
import { habitApi } from "../api/habitApi";
import { ApiError, errorMessage } from "../api/client";
import type { Habit, HabitEntry, HabitFrequency } from "../types/habit";
import { localDateString, startOfWeek } from "../utils/format";

const habits = ref<Habit[]>([]);
/** Entries per habit: today's for daily habits, this week's for weekly ones. */
const entries = ref<Record<number, HabitEntry[]>>({});
const isLoading = ref(true);
const error = ref<string | null>(null);

const name = ref("");
const description = ref("");
const frequency = ref<HabitFrequency>("DAILY");
const isSubmitting = ref(false);

onMounted(async () => {
  try {
    const [allHabits, todays] = await Promise.all([habitApi.getAll(), habitApi.getTodaysEntries()]);
    habits.value = allHabits;
    const byHabit: Record<number, HabitEntry[]> = {};
    for (const entry of todays) {
      (byHabit[entry.habitId] ??= []).push(entry);
    }
    // Weekly habits count as done for the whole week, so their entries are needed too.
    const weekly = allHabits.filter((h) => h.frequency === "WEEKLY");
    const weeklyEntries = await Promise.all(weekly.map((h) => habitApi.getEntries(h.id)));
    weekly.forEach((h, i) => {
      byHabit[h.id] = weeklyEntries[i];
    });
    entries.value = byHabit;
  } catch (err) {
    error.value = errorMessage(err, "Habits konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
});

function isDone(habit: Habit): boolean {
  const list = entries.value[habit.id] ?? [];
  if (habit.frequency === "WEEKLY") {
    const monday = startOfWeek();
    return list.some((e) => e.date >= monday);
  }
  const today = localDateString();
  return list.some((e) => e.date === today);
}

async function handleCreate() {
  isSubmitting.value = true;
  error.value = null;
  try {
    const habit = await habitApi.create({
      name: name.value,
      description: description.value || undefined,
      frequency: frequency.value,
    });
    habits.value.push(habit);
    name.value = "";
    description.value = "";
    frequency.value = "DAILY";
  } catch (err) {
    error.value = errorMessage(err, "Habit konnte nicht erstellt werden.");
  } finally {
    isSubmitting.value = false;
  }
}

async function handleComplete(habit: Habit) {
  error.value = null;
  try {
    const entry = await habitApi.completeToday(habit.id);
    entries.value = { ...entries.value, [habit.id]: [...(entries.value[habit.id] ?? []), entry] };
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      error.value =
        habit.frequency === "WEEKLY"
          ? "Diesen Habit hast du diese Woche bereits abgehakt."
          : "Diesen Habit hast du heute bereits abgehakt.";
    } else {
      error.value = "Habit konnte nicht abgehakt werden.";
    }
  }
}

async function handleDelete(id: number) {
  try {
    await habitApi.delete(id);
    habits.value = habits.value.filter((h) => h.id !== id);
    const rest = { ...entries.value };
    delete rest[id];
    entries.value = rest;
  } catch (err) {
    error.value = errorMessage(err, "Habit konnte nicht gelöscht werden.");
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-8">Habits</h1>

    <form class="border border-border p-5 mb-10 space-y-3" @submit.prevent="handleCreate">
      <div>
        <label for="name" class="block text-sm mb-1">Name</label>
        <input id="name" v-model="name" required maxlength="255" class="field w-full" />
      </div>
      <div>
        <label for="description" class="block text-sm mb-1">Beschreibung</label>
        <textarea id="description" v-model="description" rows="2" class="field w-full" />
      </div>
      <div>
        <label for="frequency" class="block text-sm mb-1">Häufigkeit</label>
        <select id="frequency" v-model="frequency" class="field">
          <option value="DAILY">Täglich</option>
          <option value="WEEKLY">Wöchentlich</option>
        </select>
      </div>
      <button type="submit" :disabled="isSubmitting" class="btn-primary">
        {{ isSubmitting ? "Speichert…" : "Habit hinzufügen" }}
      </button>
    </form>

    <p v-if="error" class="text-sm text-red-700 mb-4">{{ error }}</p>
    <p v-if="isLoading" class="text-ink-muted">Lädt…</p>
    <p v-else-if="habits.length === 0" class="text-sm text-ink-muted">Noch keine Habits angelegt.</p>

    <ul v-if="habits.length > 0" class="divide-y divide-border border-t border-b border-border">
      <li v-for="habit in habits" :key="habit.id" class="py-4 flex items-start justify-between gap-4">
        <div>
          <p :class="{ 'line-through text-ink-muted': isDone(habit) }">{{ habit.name }}</p>
          <p v-if="habit.description" class="text-sm text-ink-muted mt-1">{{ habit.description }}</p>
          <p class="text-xs text-ink-muted mt-1">{{ habit.frequency === "DAILY" ? "Täglich" : "Wöchentlich" }}</p>
        </div>
        <div class="flex gap-3 shrink-0">
          <button v-if="!isDone(habit)" class="text-sm text-accent underline" @click="handleComplete(habit)">
            {{ habit.frequency === "WEEKLY" ? "Diese Woche abhaken" : "Heute abhaken" }}
          </button>
          <span v-else class="text-sm text-ink-muted">
            ✓ {{ habit.frequency === "WEEKLY" ? "Diese Woche erledigt" : "Heute erledigt" }}
          </span>
          <button class="text-sm text-ink-muted hover:text-red-700 underline" @click="handleDelete(habit.id)">
            Löschen
          </button>
        </div>
      </li>
    </ul>
  </div>
</template>
