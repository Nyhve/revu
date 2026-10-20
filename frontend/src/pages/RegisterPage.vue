<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { authApi } from "../api/authApi";
import { errorMessage } from "../api/client";
import { APP_NAME } from "../config";

const email = ref("");
const password = ref("");
const error = ref<string | null>(null);
const isSubmitting = ref(false);
const router = useRouter();

async function handleSubmit() {
  error.value = null;
  isSubmitting.value = true;
  try {
    await authApi.register({ email: email.value, password: password.value });
    await router.push({ name: "login" });
  } catch (err) {
    error.value = errorMessage(err, "Registrierung fehlgeschlagen. Bitte versuch es erneut.");
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-semibold mb-1">Konto erstellen</h1>
      <p class="text-ink-muted mb-8">Leg deinen {{ APP_NAME }}-Account an.</p>

      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div>
          <label for="email" class="block text-sm mb-1">Email</label>
          <input id="email" v-model="email" type="email" autocomplete="email" required class="field w-full" />
        </div>
        <div>
          <label for="password" class="block text-sm mb-1">Passwort</label>
          <input id="password" v-model="password" type="password" autocomplete="new-password" required minlength="8" class="field w-full" />
          <p class="text-xs text-ink-muted mt-1">Mindestens 8 Zeichen.</p>
        </div>

        <p v-if="error" class="text-sm text-red-700">{{ error }}</p>

        <button type="submit" :disabled="isSubmitting" class="btn-primary w-full">
          {{ isSubmitting ? "Erstellt Konto…" : "Registrieren" }}
        </button>
      </form>

      <p class="text-sm text-ink-muted mt-6">
        Schon einen Account?
        <RouterLink to="/login" class="text-accent underline">Anmelden</RouterLink>
      </p>
    </div>
  </div>
</template>
