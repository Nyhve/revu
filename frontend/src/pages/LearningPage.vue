<script setup lang="ts">
import { onMounted, ref } from "vue";
import { learningApi } from "../api/learningApi";
import { errorMessage } from "../api/client";
import type { LearningProject, LearningSession } from "../types/learning";
import { formatMinutes, localDateString } from "../utils/format";

const projects = ref<LearningProject[]>([]);
const sessions = ref<Record<number, LearningSession[]>>({});
const expandedProjectId = ref<number | null>(null);
const addingSessionFor = ref<number | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);

const title = ref("");
const description = ref("");
const topic = ref("");

const sessionDuration = ref("");
const sessionNotes = ref("");
const sessionDate = ref(localDateString());

onMounted(async () => {
  try {
    projects.value = await learningApi.getAll();
  } catch (err) {
    error.value = errorMessage(err, "Projekte konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
});

function changeTotal(projectId: number, delta: number) {
  projects.value = projects.value.map((p) =>
    p.id === projectId ? { ...p, totalMinutes: p.totalMinutes + delta } : p,
  );
}

async function handleCreateProject() {
  error.value = null;
  try {
    const project = await learningApi.create({
      title: title.value,
      description: description.value || undefined,
      topic: topic.value,
    });
    projects.value = [project, ...projects.value];
    title.value = "";
    description.value = "";
    topic.value = "";
  } catch (err) {
    error.value = errorMessage(err, "Projekt konnte nicht erstellt werden.");
  }
}

async function handleDeleteProject(id: number) {
  try {
    await learningApi.delete(id);
    projects.value = projects.value.filter((p) => p.id !== id);
    const rest = { ...sessions.value };
    delete rest[id];
    sessions.value = rest;
    if (expandedProjectId.value === id) expandedProjectId.value = null;
  } catch (err) {
    error.value = errorMessage(err, "Projekt konnte nicht gelöscht werden.");
  }
}

async function toggleProject(id: number) {
  if (expandedProjectId.value === id) {
    expandedProjectId.value = null;
    return;
  }
  expandedProjectId.value = id;
  if (!sessions.value[id]) {
    try {
      sessions.value = { ...sessions.value, [id]: await learningApi.getSessions(id) };
    } catch (err) {
      error.value = errorMessage(err, "Sessions konnten nicht geladen werden.");
    }
  }
}

async function handleAddSession(projectId: number) {
  const minutes = parseInt(sessionDuration.value, 10);
  if (!minutes) return;
  error.value = null;
  try {
    const session = await learningApi.addSession(projectId, {
      durationMinutes: minutes,
      notes: sessionNotes.value || undefined,
      date: sessionDate.value,
    });
    if (sessions.value[projectId]) {
      sessions.value = { ...sessions.value, [projectId]: [...sessions.value[projectId], session] };
    }
    changeTotal(projectId, minutes);
    sessionDuration.value = "";
    sessionNotes.value = "";
    addingSessionFor.value = null;
  } catch (err) {
    error.value = errorMessage(err, "Session konnte nicht hinzugefügt werden.");
  }
}

async function handleDeleteSession(projectId: number, session: LearningSession) {
  try {
    await learningApi.deleteSession(projectId, session.id);
    sessions.value = {
      ...sessions.value,
      [projectId]: (sessions.value[projectId] ?? []).filter((s) => s.id !== session.id),
    };
    changeTotal(projectId, -session.durationMinutes);
  } catch (err) {
    error.value = errorMessage(err, "Session konnte nicht gelöscht werden.");
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-8">Learning</h1>

    <form class="border border-border p-5 mb-10 space-y-3" @submit.prevent="handleCreateProject">
      <div>
        <label for="title" class="block text-sm mb-1">Titel</label>
        <input id="title" v-model="title" required maxlength="255" class="field w-full" />
      </div>
      <div>
        <label for="topic" class="block text-sm mb-1">Thema</label>
        <input id="topic" v-model="topic" required maxlength="255" class="field w-full" />
      </div>
      <div>
        <label for="description" class="block text-sm mb-1">Beschreibung</label>
        <textarea id="description" v-model="description" rows="2" class="field w-full" />
      </div>
      <button type="submit" class="btn-primary">Projekt anlegen</button>
    </form>

    <p v-if="error" class="text-sm text-red-700 mb-4">{{ error }}</p>
    <p v-if="isLoading" class="text-ink-muted">Lädt…</p>
    <p v-else-if="projects.length === 0" class="text-sm text-ink-muted">Noch keine Lernprojekte angelegt.</p>

    <ul v-if="projects.length > 0" class="divide-y divide-border border-t border-b border-border">
      <li v-for="project in projects" :key="project.id">
        <div class="py-4 flex items-start justify-between gap-4">
          <div>
            <button
              class="font-medium hover:text-accent text-left"
              :aria-expanded="expandedProjectId === project.id"
              @click="toggleProject(project.id)"
            >
              {{ project.title }}
            </button>
            <p class="text-sm text-ink-muted mt-1">{{ project.topic }} · {{ formatMinutes(project.totalMinutes) }}</p>
            <p v-if="project.description" class="text-sm text-ink-muted">{{ project.description }}</p>
          </div>
          <div class="flex gap-3 shrink-0">
            <button
              class="text-sm text-accent underline"
              @click="addingSessionFor = addingSessionFor === project.id ? null : project.id"
            >
              {{ addingSessionFor === project.id ? "Abbrechen" : "Session eintragen" }}
            </button>
            <button class="text-sm text-ink-muted hover:text-red-700 underline" @click="handleDeleteProject(project.id)">
              Löschen
            </button>
          </div>
        </div>

        <form v-if="addingSessionFor === project.id" class="pb-4 flex gap-2 flex-wrap" @submit.prevent="handleAddSession(project.id)">
          <input v-model="sessionDuration" type="number" min="1" placeholder="Minuten" required aria-label="Minuten" class="field w-28 text-sm py-1" />
          <input v-model="sessionDate" type="date" required aria-label="Datum" class="field text-sm py-1" />
          <input v-model="sessionNotes" placeholder="Notizen (optional)" aria-label="Notizen" class="field flex-1 min-w-40 text-sm py-1" />
          <button type="submit" class="btn-primary text-sm py-1">Speichern</button>
        </form>

        <div v-if="expandedProjectId === project.id" class="pb-4">
          <p v-if="!sessions[project.id] || sessions[project.id].length === 0" class="text-sm text-ink-muted">
            Noch keine Sessions.
          </p>
          <ul v-else class="space-y-2">
            <li
              v-for="session in sessions[project.id]"
              :key="session.id"
              class="flex items-center justify-between text-sm border-l-2 border-border pl-3"
            >
              <div>
                <span class="text-ink-muted">{{ session.date }}</span>
                · <span>{{ formatMinutes(session.durationMinutes) }}</span>
                <span v-if="session.notes" class="text-ink-muted"> · {{ session.notes }}</span>
              </div>
              <button class="text-xs text-ink-muted hover:text-red-700 underline" @click="handleDeleteSession(project.id, session)">
                Löschen
              </button>
            </li>
          </ul>
        </div>
      </li>
    </ul>
  </div>
</template>
