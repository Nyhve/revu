<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { authApi } from "../api/authApi";
import { errorMessage } from "../api/client";
import { useAuthStore } from "../stores/auth";
import { APP_NAME } from "../config";

const email = ref("");
const password = ref("");
const error = ref<string | null>(null);
const isSubmitting = ref(false);

const auth = useAuthStore();
const router = useRouter();

async function handleSubmit() {
  error.value = null;
  isSubmitting.value = true;
  try {
    const response = await authApi.login({ email: email.value, password: password.value });
    auth.login(response.token);
    await router.push({ name: "dashboard" });
  } catch (err) {
    error.value = errorMessage(err, "Anmeldung fehlgeschlagen. Bitte versuch es erneut.");
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-semibold mb-1">{{ APP_NAME }}</h1>
      <p class="text-ink-muted mb-8">Melde dich an, um weiterzumachen.</p>

      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div>
          <label for="email" class="block text-sm mb-1">Email</label>
          <input id="email" v-model="email" type="email" autocomplete="email" required class="field w-full" />
        </div>
        <div>
          <label for="password" class="block text-sm mb-1">Passwort</label>
          <input id="password" v-model="password" type="password" autocomplete="current-password" required class="field w-full" />
        </div>

        <p v-if="error" class="text-sm text-red-700">{{ error }}</p>

        <button type="submit" :disabled="isSubmitting" class="btn-primary w-full">
          {{ isSubmitting ? "Meldet an…" : "Anmelden" }}
        </button>
      </form>

      <p class="text-sm text-ink-muted mt-6">
        Noch keinen Account?
        <RouterLink to="/register" class="text-accent underline">Registrieren</RouterLink>
      </p>
    </div>
  </div>
</template>
