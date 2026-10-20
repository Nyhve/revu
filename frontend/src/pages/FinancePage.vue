<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { financeApi } from "../api/financeApi";
import { errorMessage } from "../api/client";
import type { MonthlyStats, Transaction, TransactionCategory, TransactionType } from "../types/finance";
import { formatEuro, localDateString } from "../utils/format";

const CATEGORY_LABELS: Record<TransactionCategory, string> = {
  SALARY: "Gehalt",
  FREELANCE: "Freiberuflich",
  OTHER_INCOME: "Sonstige Einnahmen",
  FOOD: "Lebensmittel",
  TRANSPORT: "Transport",
  HOUSING: "Wohnen",
  HEALTH: "Gesundheit",
  ENTERTAINMENT: "Freizeit",
  SHOPPING: "Einkaufen",
  EDUCATION: "Bildung",
  OTHER_EXPENSE: "Sonstige Ausgaben",
};
const INCOME_CATEGORIES: TransactionCategory[] = ["SALARY", "FREELANCE", "OTHER_INCOME"];
const EXPENSE_CATEGORIES: TransactionCategory[] = [
  "FOOD", "TRANSPORT", "HOUSING", "HEALTH", "ENTERTAINMENT", "SHOPPING", "EDUCATION", "OTHER_EXPENSE",
];
const MONTH_NAMES = [
  "Januar", "Februar", "März", "April", "Mai", "Juni",
  "Juli", "August", "September", "Oktober", "November", "Dezember",
];

const now = new Date();
const years = Array.from({ length: 5 }, (_, i) => now.getFullYear() - 3 + i);

const transactions = ref<Transaction[]>([]);
const stats = ref<MonthlyStats | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);

const selectedYear = ref(now.getFullYear());
const selectedMonth = ref(now.getMonth() + 1);

const amount = ref("");
const description = ref("");
const type = ref<TransactionType>("EXPENSE");
const category = ref<TransactionCategory>("FOOD");
const date = ref(localDateString(now));
const isSubmitting = ref(false);

const availableCategories = computed(() => (type.value === "INCOME" ? INCOME_CATEGORIES : EXPENSE_CATEGORIES));
const selectedPrefix = computed(() => `${selectedYear.value}-${String(selectedMonth.value).padStart(2, "0")}`);

async function loadStats() {
  try {
    stats.value = await financeApi.getMonthlyStats(selectedYear.value, selectedMonth.value);
  } catch (err) {
    error.value = errorMessage(err, "Statistik konnte nicht geladen werden.");
  }
}

onMounted(async () => {
  try {
    const [all] = await Promise.all([financeApi.getAll(), loadStats()]);
    transactions.value = all;
  } catch (err) {
    error.value = errorMessage(err, "Daten konnten nicht geladen werden.");
  } finally {
    isLoading.value = false;
  }
});

watch([selectedYear, selectedMonth], loadStats);

function setType(newType: TransactionType) {
  type.value = newType;
  category.value = newType === "INCOME" ? "SALARY" : "FOOD";
}

async function handleCreate() {
  isSubmitting.value = true;
  error.value = null;
  try {
    const transaction = await financeApi.create({
      amount: parseFloat(amount.value),
      description: description.value,
      category: category.value,
      type: type.value,
      date: date.value,
    });
    transactions.value = [transaction, ...transactions.value].sort((a, b) => b.date.localeCompare(a.date));
    if (transaction.date.startsWith(selectedPrefix.value)) {
      await loadStats();
    }
    amount.value = "";
    description.value = "";
  } catch (err) {
    error.value = errorMessage(err, "Transaktion konnte nicht erstellt werden.");
  } finally {
    isSubmitting.value = false;
  }
}

async function handleDelete(id: number) {
  try {
    await financeApi.delete(id);
    transactions.value = transactions.value.filter((t) => t.id !== id);
    await loadStats();
  } catch (err) {
    error.value = errorMessage(err, "Transaktion konnte nicht gelöscht werden.");
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-8">Finanzen</h1>

    <p v-if="error" class="text-sm text-red-700 mb-4">{{ error }}</p>

    <form class="border border-border p-5 mb-8 space-y-3" @submit.prevent="handleCreate">
      <div class="flex gap-3">
        <button
          v-for="option in (['EXPENSE', 'INCOME'] as const)"
          :key="option"
          type="button"
          class="flex-1 rounded px-3 py-2 text-sm border"
          :class="type === option ? 'bg-accent text-white border-accent' : 'border-border'"
          :aria-pressed="type === option"
          @click="setType(option)"
        >
          {{ option === "EXPENSE" ? "Ausgabe" : "Einnahme" }}
        </button>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label for="amount" class="block text-sm mb-1">Betrag (€)</label>
          <input id="amount" v-model="amount" type="number" step="0.01" min="0.01" required class="field w-full" />
        </div>
        <div>
          <label for="date" class="block text-sm mb-1">Datum</label>
          <input id="date" v-model="date" type="date" required class="field w-full" />
        </div>
      </div>
      <div>
        <label for="description" class="block text-sm mb-1">Beschreibung</label>
        <input id="description" v-model="description" required maxlength="255" class="field w-full" />
      </div>
      <div>
        <label for="category" class="block text-sm mb-1">Kategorie</label>
        <select id="category" v-model="category" class="field w-full">
          <option v-for="cat in availableCategories" :key="cat" :value="cat">{{ CATEGORY_LABELS[cat] }}</option>
        </select>
      </div>
      <button type="submit" :disabled="isSubmitting" class="btn-primary">
        {{ isSubmitting ? "Speichert…" : "Hinzufügen" }}
      </button>
    </form>

    <div class="flex items-center gap-3 mb-6">
      <select v-model.number="selectedMonth" class="field text-sm" aria-label="Monat">
        <option v-for="(monthName, i) in MONTH_NAMES" :key="i" :value="i + 1">{{ monthName }}</option>
      </select>
      <select v-model.number="selectedYear" class="field text-sm" aria-label="Jahr">
        <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
      </select>
    </div>

    <div v-if="stats" class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
      <div class="border border-border p-4">
        <p class="text-2xl font-semibold text-green-700">{{ formatEuro(stats.totalIncome) }}</p>
        <p class="text-sm text-ink-muted mt-1">Einnahmen</p>
      </div>
      <div class="border border-border p-4">
        <p class="text-2xl font-semibold text-red-700">{{ formatEuro(stats.totalExpenses) }}</p>
        <p class="text-sm text-ink-muted mt-1">Ausgaben</p>
      </div>
      <div class="border border-border p-4">
        <p class="text-2xl font-semibold" :class="stats.balance >= 0 ? 'text-green-700' : 'text-red-700'">
          {{ formatEuro(stats.balance) }}
        </p>
        <p class="text-sm text-ink-muted mt-1">Bilanz</p>
      </div>
    </div>

    <p v-if="isLoading" class="text-ink-muted">Lädt…</p>
    <p v-else-if="transactions.length === 0" class="text-sm text-ink-muted">Noch keine Transaktionen.</p>

    <ul v-if="transactions.length > 0" class="divide-y divide-border border-t border-b border-border">
      <li v-for="t in transactions" :key="t.id" class="py-3 flex items-center justify-between gap-4">
        <div>
          <p>{{ t.description }}</p>
          <p class="text-xs text-ink-muted mt-1">{{ CATEGORY_LABELS[t.category] }} · {{ t.date }}</p>
        </div>
        <div class="flex items-center gap-4 shrink-0">
          <span class="font-medium" :class="t.type === 'INCOME' ? 'text-green-700' : 'text-red-700'">
            {{ t.type === "INCOME" ? "+" : "-" }}{{ formatEuro(t.amount) }}
          </span>
          <button class="text-sm text-ink-muted hover:text-red-700 underline" @click="handleDelete(t.id)">Löschen</button>
        </div>
      </li>
    </ul>
  </div>
</template>
