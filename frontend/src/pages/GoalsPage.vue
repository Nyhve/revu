<script setup lang="ts">
import { onMounted, ref } from "vue";
import { goalApi } from "../api/goalApi";
import { errorMessage } from "../api/client";
import type { Goal, GoalStatus } from "../types/goal";

const STATUS_LABELS: Record<GoalStatus, string> = {
  ACTIVE: "Aktiv",
  COMPLETED: "Erledigt",
  ARCHIVED: "Archiviert",
};

const goals = ref<Goal[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);

const title = ref("");
const description = ref("");
const targetDate = ref("");
const isSubmitting = ref(false);

onMounted(async () => {
  try {
    goals.value = await goalApi.getAll();
  } catch (err) {
    error.value = errorMessage(err, "Ziele konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
});

async function handleCreate() {
  isSubmitting.value = true;
  error.value = null;
  try {
    const goal = await goalApi.create({
      title: title.value,
      description: description.value || undefined,
      targetDate: targetDate.value || undefined,
    });
    goals.value.push(goal);
    title.value = "";
    description.value = "";
    targetDate.value = "";
  } catch (err) {
    error.value = errorMessage(err, "Ziel konnte nicht erstellt werden.");
  } finally {
    isSubmitting.value = false;
  }
}

async function handleComplete(goal: Goal) {
  try {
    const updated = await goalApi.update(goal.id, {
      title: goal.title,
      description: goal.description ?? undefined,
      targetDate: goal.targetDate ?? undefined,
      status: "COMPLETED",
    });
    goals.value = goals.value.map((g) => (g.id === updated.id ? updated : g));
  } catch (err) {
    error.value = errorMessage(err, "Ziel konnte nicht aktualisiert werden.");
  }
}

async function handleDelete(id: number) {
  try {
    await goalApi.delete(id);
    goals.value = goals.value.filter((g) => g.id !== id);
  } catch (err) {
    error.value = errorMessage(err, "Ziel konnte nicht gelöscht werden.");
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-8">Ziele</h1>

    <form class="border border-border p-5 mb-10 space-y-3" @submit.prevent="handleCreate">
      <div>
        <label for="title" class="block text-sm mb-1">Titel</label>
        <input id="title" v-model="title" required maxlength="255" class="field w-full" />
      </div>
      <div>
        <label for="description" class="block text-sm mb-1">Beschreibung</label>
        <textarea id="description" v-model="description" rows="2" class="field w-full" />
      </div>
      <div>
        <label for="targetDate" class="block text-sm mb-1">Zieldatum</label>
        <input id="targetDate" v-model="targetDate" type="date" class="field" />
      </div>
      <button type="submit" :disabled="isSubmitting" class="btn-primary">
        {{ isSubmitting ? "Speichert…" : "Ziel hinzufügen" }}
      </button>
    </form>

    <p v-if="error" class="text-sm text-red-700 mb-4">{{ error }}</p>
    <p v-if="isLoading" class="text-ink-muted">Lädt…</p>
    <p v-else-if="goals.length === 0" class="text-sm text-ink-muted">Noch keine Ziele angelegt.</p>

    <ul v-if="goals.length > 0" class="divide-y divide-border border-t border-b border-border">
      <li v-for="goal in goals" :key="goal.id" class="py-4 flex items-start justify-between gap-4">
        <div>
          <p :class="{ 'line-through text-ink-muted': goal.status === 'COMPLETED' }">{{ goal.title }}</p>
          <p v-if="goal.description" class="text-sm text-ink-muted mt-1">{{ goal.description }}</p>
          <p class="text-xs text-ink-muted mt-1">
            {{ STATUS_LABELS[goal.status] }}<template v-if="goal.targetDate"> · Ziel: {{ goal.targetDate }}</template>
          </p>
        </div>
        <div class="flex gap-3 shrink-0">
          <button v-if="goal.status === 'ACTIVE'" class="text-sm text-accent underline" @click="handleComplete(goal)">
            Erledigt
          </button>
          <button class="text-sm text-ink-muted hover:text-red-700 underline" @click="handleDelete(goal.id)">
            Löschen
          </button>
        </div>
      </li>
    </ul>
  </div>
</template>
