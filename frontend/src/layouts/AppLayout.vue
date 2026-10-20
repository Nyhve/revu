<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, RouterView, useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import type { UserPreferences } from "../types/user";
import { APP_NAME } from "../config";

const NAV_LINKS: { to: string; label: string; feature: keyof UserPreferences }[] = [
  { to: "/goals", label: "Ziele", feature: "goalsEnabled" },
  { to: "/tasks", label: "Aufgaben", feature: "tasksEnabled" },
  { to: "/habits", label: "Habits", feature: "habitsEnabled" },
  { to: "/fitness", label: "Fitness", feature: "fitnessEnabled" },
  { to: "/learning", label: "Learning", feature: "learningEnabled" },
  { to: "/finance", label: "Finanzen", feature: "financeEnabled" },
];

const auth = useAuthStore();
const router = useRouter();
const menuOpen = ref(false);

const visibleLinks = computed(() => NAV_LINKS.filter((link) => auth.features[link.feature]));

function logout() {
  menuOpen.value = false;
  auth.logout();
  router.replace({ name: "login" });
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <header class="border-b border-border bg-surface sticky top-0 z-50">
      <div class="max-w-3xl mx-auto px-4 md:px-8 py-4 flex items-center justify-between">
        <div class="flex items-center gap-6">
          <RouterLink to="/" class="font-semibold hover:text-accent transition-colors" @click="menuOpen = false">
            {{ APP_NAME }}
          </RouterLink>
          <nav class="hidden md:flex gap-4 text-sm">
            <RouterLink
              v-for="link in visibleLinks"
              :key="link.to"
              :to="link.to"
              class="text-ink-muted hover:text-ink"
              active-class="text-accent! font-medium"
            >
              {{ link.label }}
            </RouterLink>
          </nav>
        </div>

        <div class="hidden md:flex items-center gap-4">
          <RouterLink to="/profile" class="text-sm text-ink-muted hover:text-ink" active-class="text-accent! font-medium">
            Profil
          </RouterLink>
          <button class="text-sm text-ink-muted hover:text-ink underline" @click="logout">Abmelden</button>
        </div>

        <button
          class="md:hidden flex flex-col gap-1.5 p-2"
          :aria-label="menuOpen ? 'Menü schließen' : 'Menü öffnen'"
          :aria-expanded="menuOpen"
          @click="menuOpen = !menuOpen"
        >
          <span class="block w-5 h-0.5 bg-ink transition-transform duration-200" :class="{ 'rotate-45 translate-y-2': menuOpen }" />
          <span class="block w-5 h-0.5 bg-ink transition-opacity duration-200" :class="{ 'opacity-0': menuOpen }" />
          <span class="block w-5 h-0.5 bg-ink transition-transform duration-200" :class="{ '-rotate-45 -translate-y-2': menuOpen }" />
        </button>
      </div>

      <div v-if="menuOpen" class="md:hidden border-t border-border bg-surface">
        <RouterLink
          v-for="link in [...visibleLinks, { to: '/profile', label: 'Profil' }]"
          :key="link.to"
          :to="link.to"
          class="block py-3 px-4 text-sm border-b border-border text-ink hover:bg-surface"
          active-class="text-accent! font-medium"
          @click="menuOpen = false"
        >
          {{ link.label }}
        </RouterLink>
        <button class="block w-full text-left py-3 px-4 text-sm text-ink-muted hover:bg-surface" @click="logout">
          Abmelden
        </button>
      </div>
    </header>

    <main class="max-w-3xl mx-auto px-4 md:px-8 py-6 md:py-8">
      <RouterView />
    </main>
  </div>
</template>
