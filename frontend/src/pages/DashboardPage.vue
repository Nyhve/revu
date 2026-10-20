<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { goalApi } from "../api/goalApi";
import { taskApi } from "../api/taskApi";
import { habitApi, type HabitConsistency } from "../api/habitApi";
import { fitnessApi } from "../api/fitnessApi";
import { learningApi } from "../api/learningApi";
import { financeApi } from "../api/financeApi";
import { activityApi, type DailyActivityCount } from "../api/activityApi";
import { errorMessage } from "../api/client";
import { useAuthStore } from "../stores/auth";
import type { Goal } from "../types/goal";
import type { Task } from "../types/task";
import type { Habit } from "../types/habit";
import type { Workout } from "../types/fitness";
import type { LearningProject } from "../types/learning";
import type { MonthlyFinanceSummary, MonthlyStats } from "../types/finance";
import AppChart from "../components/charts/AppChart.vue";
import { useChartTheme } from "../components/charts/chartTheme";
import { formatEuro, isoWeek, parseLocalDate } from "../utils/format";

const auth = useAuthStore();
const chartTheme = useChartTheme();

const goals = ref<Goal[]>([]);
const tasks = ref<Task[]>([]);
const habits = ref<Habit[]>([]);
const workouts = ref<Workout[]>([]);
const learningProjects = ref<LearningProject[]>([]);
const monthlyStats = ref<MonthlyStats | null>(null);
const dailyActivity = ref<DailyActivityCount[]>([]);
const financeSummary = ref<MonthlyFinanceSummary[]>([]);
const habitConsistency = ref<HabitConsistency[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);

const f = computed(() => auth.features);

async function load() {
  const now = new Date();
  const on = f.value;
  isLoading.value = true;
  error.value = null;
  try {
    // Only modules that are switched on are loaded and shown.
    await Promise.all([
      on.goalsEnabled && goalApi.getAll().then((v) => (goals.value = v)),
      on.tasksEnabled && taskApi.getAll().then((v) => (tasks.value = v)),
      on.habitsEnabled && habitApi.getAll().then((v) => (habits.value = v)),
      on.habitsEnabled && habitApi.getConsistency(28).then((v) => (habitConsistency.value = v)),
      on.fitnessEnabled && fitnessApi.getWorkouts().then((v) => (workouts.value = v)),
      on.learningEnabled && learningApi.getAll().then((v) => (learningProjects.value = v)),
      on.financeEnabled &&
        financeApi.getMonthlyStats(now.getFullYear(), now.getMonth() + 1).then((v) => (monthlyStats.value = v)),
      on.financeEnabled && financeApi.getLastMonthsSummary(6).then((v) => (financeSummary.value = v)),
      activityApi.getDailyCounts(30).then((v) => (dailyActivity.value = v)),
    ]);
  } catch (err) {
    error.value = errorMessage(err, "Daten konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
}

// Wait until the feature toggles are known, then load (and reload if they change).
watch(
  () => auth.preferences,
  (prefs) => {
    if (prefs) void load();
  },
  { immediate: true },
);

const today = new Date().toLocaleDateString("de-DE", { weekday: "long", year: "numeric", month: "long", day: "numeric" });

const activeGoals = computed(() => goals.value.filter((g) => g.status === "ACTIVE"));
const openTasks = computed(() => tasks.value.filter((t) => !t.completed));
const upcomingTasks = computed(() =>
  [...openTasks.value]
    .sort((a, b) => {
      if (!a.dueDate) return 1;
      if (!b.dueDate) return -1;
      return a.dueDate.localeCompare(b.dueDate);
    })
    .slice(0, 5),
);
const completedWorkouts = computed(() => workouts.value.filter((w) => w.completed));
const learningMinutes = computed(() => learningProjects.value.reduce((sum, p) => sum + p.totalMinutes, 0));
const hasFinanceData = computed(() => financeSummary.value.some((m) => m.income > 0 || m.expenses > 0));

const workoutsPerWeek = computed(() => {
  const weeks = new Map<string, number>();
  const now = new Date();
  for (let i = 7; i >= 0; i--) {
    const d = new Date(now);
    d.setDate(now.getDate() - i * 7);
    weeks.set(`KW ${isoWeek(d)}`, 0);
  }
  for (const w of completedWorkouts.value) {
    const label = `KW ${isoWeek(parseLocalDate(w.date))}`;
    if (weeks.has(label)) weeks.set(label, (weeks.get(label) ?? 0) + 1);
  }
  return { labels: [...weeks.keys()], data: [...weeks.values()] };
});

const volumePerWorkout = computed(() => {
  const recent = [...completedWorkouts.value].sort((a, b) => a.date.localeCompare(b.date)).slice(-10);
  return {
    labels: recent.map((w) => w.date.slice(5)),
    data: recent.map((w) =>
      Math.round(w.exercises.reduce((sum, ex) => sum + ex.sets.reduce((s, set) => s + set.weightKg * set.reps, 0), 0)),
    ),
  };
});

const stats = computed(() => {
  const list: { value: string | number; label: string; tone?: "good" | "bad" }[] = [];
  if (f.value.goalsEnabled) list.push({ value: activeGoals.value.length, label: "Aktive Ziele" });
  if (f.value.tasksEnabled) list.push({ value: openTasks.value.length, label: "Offene Aufgaben" });
  if (f.value.habitsEnabled) list.push({ value: habits.value.length, label: "Aktive Habits" });
  if (f.value.fitnessEnabled) list.push({ value: completedWorkouts.value.length, label: "Workouts" });
  if (f.value.learningEnabled) list.push({ value: learningMinutes.value, label: "Lernminuten" });
  if (f.value.financeEnabled && monthlyStats.value) {
    list.push({
      value: formatEuro(monthlyStats.value.balance),
      label: "Bilanz diesen Monat",
      tone: monthlyStats.value.balance >= 0 ? "good" : "bad",
    });
  }
  return list;
});
</script>

<template>
  <p v-if="isLoading" class="text-ink-muted">Lädt…</p>
  <p v-else-if="error" class="text-sm text-red-700">{{ error }}</p>

  <div v-else class="space-y-10">
    <div>
      <h1 class="text-2xl font-semibold mb-1">Deine Übersicht</h1>
      <p class="text-sm text-ink-muted">{{ today }}</p>
    </div>

    <div v-if="stats.length > 0" class="grid grid-cols-2 md:grid-cols-3 gap-4">
      <div v-for="s in stats" :key="s.label" class="border border-border p-5">
        <p
          class="text-3xl font-semibold"
          :class="{ 'text-green-700': s.tone === 'good', 'text-red-700': s.tone === 'bad' }"
        >
          {{ s.value }}
        </p>
        <p class="text-sm text-ink-muted mt-1">{{ s.label }}</p>
      </div>
    </div>

    <div v-if="f.tasksEnabled || f.goalsEnabled" class="grid grid-cols-1 md:grid-cols-2 gap-6 md:gap-8">
      <section v-if="f.tasksEnabled">
        <h2 class="text-sm font-medium text-ink-muted mb-3">Nächste Aufgaben</h2>
        <p v-if="upcomingTasks.length === 0" class="text-sm text-ink-muted">Keine offenen Aufgaben.</p>
        <ul v-else class="divide-y divide-border border-t border-b border-border">
          <li v-for="task in upcomingTasks" :key="task.id" class="py-3 flex items-center justify-between gap-3">
            <span class="text-sm">{{ task.title }}</span>
            <span v-if="task.dueDate" class="text-xs text-ink-muted shrink-0">{{ task.dueDate }}</span>
          </li>
        </ul>
      </section>

      <section v-if="f.goalsEnabled">
        <h2 class="text-sm font-medium text-ink-muted mb-3">Aktive Ziele</h2>
        <p v-if="activeGoals.length === 0" class="text-sm text-ink-muted">Noch keine aktiven Ziele.</p>
        <ul v-else class="divide-y divide-border border-t border-b border-border">
          <li v-for="goal in activeGoals.slice(0, 5)" :key="goal.id" class="py-3 flex items-center justify-between gap-3">
            <span class="text-sm">{{ goal.title }}</span>
            <span v-if="goal.targetDate" class="text-xs text-ink-muted shrink-0">{{ goal.targetDate }}</span>
          </li>
        </ul>
      </section>
    </div>

    <section v-if="dailyActivity.length > 0">
      <h2 class="text-sm font-medium text-ink-muted mb-4">Aktivitäten – letzte 30 Tage</h2>
      <AppChart
        type="bar"
        description="Aktivitäten pro Tag in den letzten 30 Tagen"
        :labels="dailyActivity.map((d) => d.date.slice(5))"
        :series="[{ label: 'Aktivitäten', data: dailyActivity.map((d) => d.count) }]"
        :format-value="(v) => `${v} Aktivitäten`"
        :format-title="(l) => `Datum: ${l}`"
        integer-ticks
      />
    </section>

    <section v-if="f.habitsEnabled && habitConsistency.length > 0">
      <h2 class="text-sm font-medium text-ink-muted mb-4">Habit-Konsistenz – letzte 28 Tage</h2>
      <div class="space-y-3">
        <div v-for="h in habitConsistency" :key="h.habitId">
          <div class="flex justify-between text-sm mb-1">
            <span>{{ h.name }}</span>
            <span class="text-ink-muted">{{ h.completedDays }}/{{ h.totalDays }} · {{ h.consistencyPercent }}%</span>
          </div>
          <div
            class="h-2 bg-border rounded-full overflow-hidden"
            role="progressbar"
            :aria-valuenow="h.consistencyPercent"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-label="h.name"
          >
            <div class="h-full rounded-full" :style="{ width: `${h.consistencyPercent}%`, backgroundColor: chartTheme.accent }" />
          </div>
        </div>
      </div>
    </section>

    <section v-if="f.financeEnabled && hasFinanceData">
      <h2 class="text-sm font-medium text-ink-muted mb-4">Finanzen – letzte 6 Monate</h2>
      <AppChart
        type="line"
        description="Einnahmen und Ausgaben der letzten 6 Monate"
        :height="220"
        legend
        :labels="financeSummary.map((m) => m.monthLabel)"
        :series="[
          { label: 'Einnahmen', data: financeSummary.map((m) => m.income) },
          { label: 'Ausgaben', data: financeSummary.map((m) => m.expenses), color: 'red' },
        ]"
        :format-value="(v, s) => `${s}: ${formatEuro(v)}`"
      />
    </section>

    <section v-if="f.learningEnabled && learningProjects.length > 0">
      <h2 class="text-sm font-medium text-ink-muted mb-4">Lernzeit pro Projekt</h2>
      <AppChart
        type="bar"
        horizontal
        description="Lernzeit pro Projekt in Minuten"
        :height="Math.max(120, learningProjects.length * 45)"
        :labels="learningProjects.map((p) => p.title)"
        :series="[{ label: 'Lernzeit', data: learningProjects.map((p) => p.totalMinutes) }]"
        :format-value="(v) => `${v} Minuten`"
        :format-tick="(v) => `${v} Min.`"
      />
    </section>

    <section v-if="f.fitnessEnabled && completedWorkouts.length > 0">
      <h2 class="text-sm font-medium text-ink-muted mb-4">Fitness – letzte 8 Wochen</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
        <div>
          <p class="text-xs text-ink-muted mb-2">Trainings pro Woche</p>
          <AppChart
            type="bar"
            description="Trainings pro Kalenderwoche"
            :height="160"
            :labels="workoutsPerWeek.labels"
            :series="[{ label: 'Trainings', data: workoutsPerWeek.data }]"
            :format-value="(v) => `${v} Trainings`"
            integer-ticks
          />
        </div>
        <div>
          <p class="text-xs text-ink-muted mb-2">Volumen (kg) pro Training</p>
          <AppChart
            type="line"
            description="Trainingsvolumen in kg der letzten Trainings"
            :height="160"
            :labels="volumePerWorkout.labels"
            :series="[{ label: 'Volumen', data: volumePerWorkout.data }]"
            :format-value="(v) => `${v} kg`"
            :format-title="(l) => `Datum: ${l}`"
          />
        </div>
      </div>
    </section>
  </div>
</template>
