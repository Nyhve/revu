<script setup lang="ts">
import { onMounted, ref } from "vue";
import { taskApi } from "../api/taskApi";
import { errorMessage } from "../api/client";
import type { Task } from "../types/task";

const tasks = ref<Task[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);

const title = ref("");
const description = ref("");
const dueDate = ref("");
const isSubmitting = ref(false);

onMounted(async () => {
  try {
    tasks.value = await taskApi.getAll();
  } catch (err) {
    error.value = errorMessage(err, "Aufgaben konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
});

async function handleCreate() {
  isSubmitting.value = true;
  error.value = null;
  try {
    const task = await taskApi.create({
      title: title.value,
      description: description.value || undefined,
      dueDate: dueDate.value || undefined,
    });
    tasks.value.push(task);
    title.value = "";
    description.value = "";
    dueDate.value = "";
  } catch (err) {
    error.value = errorMessage(err, "Aufgabe konnte nicht erstellt werden.");
  } finally {
    isSubmitting.value = false;
  }
}

async function toggleComplete(task: Task) {
  try {
    const updated = await taskApi.update(task.id, {
      title: task.title,
      description: task.description ?? undefined,
      dueDate: task.dueDate ?? undefined,
      completed: !task.completed,
    });
    tasks.value = tasks.value.map((t) => (t.id === updated.id ? updated : t));
  } catch (err) {
    error.value = errorMessage(err, "Aufgabe konnte nicht aktualisiert werden.");
  }
}

async function handleDelete(id: number) {
  try {
    await taskApi.delete(id);
    tasks.value = tasks.value.filter((t) => t.id !== id);
  } catch (err) {
    error.value = errorMessage(err, "Aufgabe konnte nicht gelöscht werden.");
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-8">Aufgaben</h1>

    <form class="border border-border p-5 mb-10 space-y-3" @submit.prevent="handleCreate">
      <div>
        <label for="title" class="block text-sm mb-1">Titel</label>
        <input id="title" v-model="title" required maxlength="255" class="field w-full" />
      </div>
      <div>
        <label for="description" class="block text-sm mb-1">Tätigkeit</label>
        <textarea id="description" v-model="description" rows="2" class="field w-full" />
      </div>
      <div>
        <label for="dueDate" class="block text-sm mb-1">Fälligkeitsdatum</label>
        <input id="dueDate" v-model="dueDate" type="date" class="field" />
      </div>
      <button type="submit" :disabled="isSubmitting" class="btn-primary">
        {{ isSubmitting ? "Speichert…" : "Aufgabe hinzufügen" }}
      </button>
    </form>

    <p v-if="error" class="text-sm text-red-700 mb-4">{{ error }}</p>
    <p v-if="isLoading" class="text-ink-muted">Lädt…</p>
    <p v-else-if="tasks.length === 0" class="text-sm text-ink-muted">Noch keine Aufgaben angelegt.</p>

    <ul v-if="tasks.length > 0" class="divide-y divide-border border-t border-b border-border">
      <li v-for="task in tasks" :key="task.id" class="py-4 flex items-start justify-between gap-4">
        <div class="flex items-start gap-3">
          <input
            type="checkbox"
            :checked="task.completed"
            class="mt-1.5 accent-accent"
            :aria-label="`${task.title} ${task.completed ? 'als offen markieren' : 'als erledigt markieren'}`"
            @change="toggleComplete(task)"
          />
          <div>
            <p :class="{ 'line-through text-ink-muted': task.completed }">{{ task.title }}</p>
            <p v-if="task.description" class="text-sm text-ink-muted mt-1">{{ task.description }}</p>
            <p v-if="task.dueDate" class="text-xs text-ink-muted mt-1">Fällig: {{ task.dueDate }}</p>
          </div>
        </div>
        <button class="text-sm text-ink-muted hover:text-red-700 underline shrink-0" @click="handleDelete(task.id)">
          Löschen
        </button>
      </li>
    </ul>
  </div>
</template>
