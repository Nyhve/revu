import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
  type ChartOptions,
} from "chart.js";
import { computed } from "vue";
import { useThemeStore } from "../../stores/theme";

ChartJS.register(CategoryScale, LinearScale, BarElement, LineElement, PointElement, Tooltip, Legend);

export const ACCENT = "#3A5A40";
export const ACCENT_DARK = "#6F9A73";
export const RED = "#B91C1C";
export const RED_DARK = "#E0686B";

/** Colours and base options that follow light/dark mode. */
export function useChartTheme() {
  const theme = useThemeStore();
  return computed(() => {
    const dark = theme.isDark;
    const grid = dark ? "#38362F" : "#E4E1DA";
    const muted = dark ? "#9B9690" : "#6B6862";
    const base: ChartOptions<"bar" | "line"> = {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      plugins: { legend: { display: false } },
      scales: {
        x: { grid: { color: grid }, border: { dash: [3, 3] }, ticks: { color: muted, font: { size: 10 } } },
        y: { grid: { color: grid }, border: { dash: [3, 3] }, ticks: { color: muted, font: { size: 10 } }, beginAtZero: true },
      },
    };
    return {
      accent: dark ? ACCENT_DARK : ACCENT,
      red: dark ? RED_DARK : RED,
      muted,
      grid,
      base,
    };
  });
}
