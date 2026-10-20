import { defineStore } from "pinia";
import { ref, watch } from "vue";

const STORAGE_KEY = "theme";

function initialIsDark(): boolean {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) return stored === "dark";
  } catch {
    // storage unavailable - fall back to the OS setting
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export const useThemeStore = defineStore("theme", () => {
  const isDark = ref(initialIsDark());

  watch(
    isDark,
    (dark) => {
      document.documentElement.classList.toggle("dark", dark);
      try {
        localStorage.setItem(STORAGE_KEY, dark ? "dark" : "light");
      } catch {
        // ignore - preference simply is not remembered
      }
    },
    { immediate: true },
  );

  function toggle(): void {
    isDark.value = !isDark.value;
  }

  return { isDark, toggle };
});
