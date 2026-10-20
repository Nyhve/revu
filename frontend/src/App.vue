<script setup lang="ts">
import { RouterView } from "vue-router";
import { useThemeStore } from "./stores/theme";

// Instantiating the store keeps <html class="dark"> in sync on every page.
useThemeStore();
</script>

<template>
  <RouterView />
</template>
