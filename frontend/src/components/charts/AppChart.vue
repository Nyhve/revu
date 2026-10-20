<script setup lang="ts">
import { computed } from "vue";
import { Bar, Line } from "vue-chartjs";
import type { ChartData, ChartOptions, TooltipItem } from "chart.js";
import { useChartTheme } from "./chartTheme";

export interface Series {
  label: string;
  data: number[];
  /** "accent" (default) or "red" */
  color?: "accent" | "red";
}

const props = withDefaults(
  defineProps<{
    type: "bar" | "line";
    labels: string[];
    series: Series[];
    height?: number;
    horizontal?: boolean;
    legend?: boolean;
    /** Formats values in tooltips. */
    formatValue?: (value: number, series: string) => string;
    /** Formats ticks on the value axis. */
    formatTick?: (value: number) => string;
    /** Formats the tooltip title (category label). */
    formatTitle?: (label: string) => string;
    integerTicks?: boolean;
    description: string;
  }>(),
  { height: 180, horizontal: false, legend: false, integerTicks: false },
);

const theme = useChartTheme();

const data = computed(() => ({
  labels: props.labels,
  datasets: props.series.map((s) => {
    const color = s.color === "red" ? theme.value.red : theme.value.accent;
    return {
      label: s.label,
      data: s.data,
      backgroundColor: color,
      borderColor: color,
      borderWidth: 2,
      borderRadius: 3,
      pointRadius: 3,
      tension: 0.3,
      maxBarThickness: 32,
    };
  }),
}));

const options = computed(() => {
  const t = theme.value;
  const valueAxis = props.horizontal ? "x" : "y";
  const scales = structuredClone(t.base.scales) as Record<string, Record<string, unknown>>;
  const valueTicks = scales[valueAxis].ticks as Record<string, unknown>;
  if (props.integerTicks) valueTicks.precision = 0;
  if (props.formatTick) {
    const fmt = props.formatTick;
    valueTicks.callback = (v: string | number) => fmt(Number(v));
  }
  if (props.horizontal) {
    scales.x.beginAtZero = true;
    scales.y.beginAtZero = false;
  }
  return {
    ...t.base,
    indexAxis: props.horizontal ? "y" : "x",
    scales,
    plugins: {
      legend: { display: props.legend, labels: { color: t.muted, boxWidth: 12, font: { size: 11 } } },
      tooltip: {
        callbacks: {
          title: (items: TooltipItem<"bar" | "line">[]) =>
            props.formatTitle ? props.formatTitle(items[0]?.label ?? "") : (items[0]?.label ?? ""),
          label: (item: TooltipItem<"bar" | "line">) => {
            const value = Number(item.raw);
            return props.formatValue ? props.formatValue(value, item.dataset.label ?? "") : String(value);
          },
        },
      },
    },
  };
});
</script>

<template>
  <div :style="{ height: `${height}px` }" role="img" :aria-label="description">
    <Bar
      v-if="type === 'bar'"
      :data="data as ChartData<'bar'>"
      :options="options as ChartOptions<'bar'>"
    />
    <Line
      v-else
      :data="data as ChartData<'line'>"
      :options="options as ChartOptions<'line'>"
    />
  </div>
</template>
