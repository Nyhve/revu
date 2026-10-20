<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { fitnessApi } from "../api/fitnessApi";
import { errorMessage } from "../api/client";
import type { Exercise, MuscleGroup, Workout } from "../types/fitness";
import { localDateString } from "../utils/format";

const MUSCLE_GROUP_LABELS: Record<MuscleGroup, string> = {
  CHEST: "Brust",
  BACK: "Rücken",
  LEGS: "Beine",
  SHOULDERS: "Schultern",
  ARMS: "Arme",
  CORE: "Core",
  FULL_BODY: "Ganzkörper",
};

const workouts = ref<Workout[]>([]);
const exercises = ref<Exercise[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const activeWorkoutId = ref<number | null>(null);

const workoutName = ref("");
const workoutDate = ref(localDateString());

const selectedExerciseId = ref<number | "">("");
const newExerciseName = ref("");
const newExerciseMuscle = ref<MuscleGroup>("CHEST");

const setWeight = ref("");
const setReps = ref("");
const activeExerciseId = ref<number | null>(null);

const activeWorkout = computed(() => workouts.value.find((w) => w.id === activeWorkoutId.value) ?? null);

onMounted(async () => {
  try {
    const [w, e] = await Promise.all([fitnessApi.getWorkouts(), fitnessApi.getExercises()]);
    workouts.value = w;
    exercises.value = e;
    activeWorkoutId.value = w.find((wo) => !wo.completed)?.id ?? null;
  } catch (err) {
    error.value = errorMessage(err, "Daten konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
});

function replaceWorkout(updated: Workout) {
  workouts.value = workouts.value.map((w) => (w.id === updated.id ? updated : w));
}

async function run(action: () => Promise<void>, fallback: string) {
  error.value = null;
  try {
    await action();
  } catch (err) {
    error.value = errorMessage(err, fallback);
  }
}

function handleCreateWorkout() {
  return run(async () => {
    const workout = await fitnessApi.createWorkout({ name: workoutName.value, date: workoutDate.value });
    workouts.value = [workout, ...workouts.value];
    activeWorkoutId.value = workout.id;
    workoutName.value = "";
  }, "Workout konnte nicht erstellt werden.");
}

function handleAddExercise() {
  const workoutId = activeWorkoutId.value;
  const exerciseId = selectedExerciseId.value;
  if (!workoutId || exerciseId === "") return;
  return run(async () => {
    replaceWorkout(await fitnessApi.addExercise(workoutId, exerciseId));
  }, "Übung konnte nicht hinzugefügt werden.");
}

function handleCreateExercise() {
  return run(async () => {
    const exercise = await fitnessApi.createExercise(newExerciseName.value, newExerciseMuscle.value);
    exercises.value = [...exercises.value, exercise];
    selectedExerciseId.value = exercise.id;
    newExerciseName.value = "";
  }, "Übung konnte nicht angelegt werden.");
}

function handleAddSet(workoutExerciseId: number) {
  const workoutId = activeWorkoutId.value;
  if (!workoutId || !setWeight.value || !setReps.value) return;
  return run(async () => {
    const updated = await fitnessApi.addSet(workoutId, workoutExerciseId, {
      weightKg: parseFloat(setWeight.value),
      reps: parseInt(setReps.value, 10),
    });
    replaceWorkout(updated);
    setWeight.value = "";
    setReps.value = "";
    activeExerciseId.value = null;
  }, "Satz konnte nicht hinzugefügt werden.");
}

function handleCompleteWorkout() {
  const workoutId = activeWorkoutId.value;
  if (!workoutId) return;
  return run(async () => {
    replaceWorkout(await fitnessApi.completeWorkout(workoutId));
    activeWorkoutId.value = null;
  }, "Workout konnte nicht abgeschlossen werden.");
}

function handleDeleteWorkout(id: number) {
  return run(async () => {
    await fitnessApi.deleteWorkout(id);
    workouts.value = workouts.value.filter((w) => w.id !== id);
    if (activeWorkoutId.value === id) activeWorkoutId.value = null;
  }, "Workout konnte nicht gelöscht werden.");
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-8">Fitness</h1>

    <p v-if="error" class="text-sm text-red-700 mb-4">{{ error }}</p>
    <p v-if="isLoading" class="text-ink-muted">Lädt…</p>

    <div v-else class="space-y-10">
      <form v-if="!activeWorkout" class="border border-border p-5 space-y-3" @submit.prevent="handleCreateWorkout">
        <h2 class="text-sm font-medium text-ink-muted">Neues Workout starten</h2>
        <div>
          <label for="wName" class="block text-sm mb-1">Name</label>
          <input id="wName" v-model="workoutName" required maxlength="255" class="field w-full" />
        </div>
        <div>
          <label for="wDate" class="block text-sm mb-1">Datum</label>
          <input id="wDate" v-model="workoutDate" type="date" required class="field" />
        </div>
        <button type="submit" class="btn-primary">Workout starten</button>
      </form>

      <div v-else class="border border-border p-5 space-y-6">
        <div class="flex items-center justify-between gap-4">
          <div>
            <h2 class="font-medium">{{ activeWorkout.name }}</h2>
            <p class="text-sm text-ink-muted">{{ activeWorkout.date }}</p>
          </div>
          <button class="btn-primary text-sm" @click="handleCompleteWorkout">Workout abschließen</button>
        </div>

        <div class="space-y-4">
          <div v-for="we in activeWorkout.exercises" :key="we.id" class="border border-border p-4">
            <div class="flex items-center justify-between mb-3">
              <div>
                <p class="font-medium">{{ we.exercise.name }}</p>
                <p class="text-xs text-ink-muted">{{ MUSCLE_GROUP_LABELS[we.exercise.muscleGroup] }}</p>
              </div>
              <button class="text-sm text-accent underline" @click="activeExerciseId = activeExerciseId === we.id ? null : we.id">
                {{ activeExerciseId === we.id ? "Abbrechen" : "Satz hinzufügen" }}
              </button>
            </div>

            <table v-if="we.sets.length > 0" class="w-full text-sm mb-3">
              <thead>
                <tr class="text-ink-muted text-left">
                  <th class="font-normal pb-1">Satz</th>
                  <th class="font-normal pb-1">Gewicht</th>
                  <th class="font-normal pb-1">Wdh.</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="set in we.sets" :key="set.id">
                  <td>{{ set.setNumber }}</td>
                  <td>{{ set.weightKg }} kg</td>
                  <td>{{ set.reps }}</td>
                </tr>
              </tbody>
            </table>

            <form v-if="activeExerciseId === we.id" class="flex gap-2 mt-2" @submit.prevent="handleAddSet(we.id)">
              <input v-model="setWeight" type="number" step="0.25" min="0" placeholder="kg" required aria-label="Gewicht in kg" class="field w-20 text-sm px-2 py-1" />
              <input v-model="setReps" type="number" min="1" placeholder="Wdh." required aria-label="Wiederholungen" class="field w-20 text-sm px-2 py-1" />
              <button type="submit" class="btn-primary text-sm px-3 py-1">Speichern</button>
            </form>
          </div>
        </div>

        <div class="space-y-3">
          <h3 class="text-sm font-medium text-ink-muted">Übung hinzufügen</h3>
          <div class="flex gap-2">
            <select v-model="selectedExerciseId" class="field flex-1 text-sm" aria-label="Übung">
              <option value="">Übung wählen…</option>
              <option v-for="ex in exercises" :key="ex.id" :value="ex.id">
                {{ ex.name }} ({{ MUSCLE_GROUP_LABELS[ex.muscleGroup] }})
              </option>
            </select>
            <button :disabled="selectedExerciseId === ''" class="btn-primary text-sm" @click="handleAddExercise">Hinzufügen</button>
          </div>

          <form class="flex gap-2" @submit.prevent="handleCreateExercise">
            <input v-model="newExerciseName" placeholder="Neue Übung…" required maxlength="255" aria-label="Neue Übung" class="field flex-1 text-sm" />
            <select v-model="newExerciseMuscle" class="field text-sm" aria-label="Muskelgruppe">
              <option v-for="(label, value) in MUSCLE_GROUP_LABELS" :key="value" :value="value">{{ label }}</option>
            </select>
            <button type="submit" class="btn-primary text-sm">Anlegen</button>
          </form>
        </div>
      </div>

      <section v-if="workouts.length > 0">
        <h2 class="text-sm font-medium text-ink-muted mb-3">Vergangene Workouts</h2>
        <ul class="divide-y divide-border border-t border-b border-border">
          <li v-for="workout in workouts" :key="workout.id" class="py-3 flex items-center justify-between">
            <div>
              <p :class="{ 'text-ink-muted': workout.completed }">{{ workout.name }}</p>
              <p class="text-xs text-ink-muted">
                {{ workout.date }} · {{ workout.exercises.length }} Übungen · {{ workout.completed ? "Abgeschlossen" : "Aktiv" }}
              </p>
            </div>
            <button class="text-sm text-ink-muted hover:text-red-700 underline" @click="handleDeleteWorkout(workout.id)">
              Löschen
            </button>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>
