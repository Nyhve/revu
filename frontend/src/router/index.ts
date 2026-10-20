import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "../stores/auth";
import type { UserPreferences } from "../types/user";
import AppLayout from "../layouts/AppLayout.vue";

declare module "vue-router" {
  interface RouteMeta {
    requiresAuth?: boolean;
    guestOnly?: boolean;
    feature?: keyof UserPreferences;
    title?: string;
  }
}

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", name: "login", component: () => import("../pages/LoginPage.vue"), meta: { guestOnly: true, title: "Anmelden" } },
    { path: "/register", name: "register", component: () => import("../pages/RegisterPage.vue"), meta: { guestOnly: true, title: "Registrieren" } },
    {
      path: "/",
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        { path: "", name: "dashboard", component: () => import("../pages/DashboardPage.vue"), meta: { title: "Übersicht" } },
        { path: "goals", name: "goals", component: () => import("../pages/GoalsPage.vue"), meta: { feature: "goalsEnabled", title: "Ziele" } },
        { path: "tasks", name: "tasks", component: () => import("../pages/TasksPage.vue"), meta: { feature: "tasksEnabled", title: "Aufgaben" } },
        { path: "habits", name: "habits", component: () => import("../pages/HabitsPage.vue"), meta: { feature: "habitsEnabled", title: "Habits" } },
        { path: "fitness", name: "fitness", component: () => import("../pages/FitnessPage.vue"), meta: { feature: "fitnessEnabled", title: "Fitness" } },
        { path: "learning", name: "learning", component: () => import("../pages/LearningPage.vue"), meta: { feature: "learningEnabled", title: "Learning" } },
        { path: "finance", name: "finance", component: () => import("../pages/FinancePage.vue"), meta: { feature: "financeEnabled", title: "Finanzen" } },
        { path: "profile", name: "profile", component: () => import("../pages/ProfilePage.vue"), meta: { title: "Profil" } },
      ],
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});

router.beforeEach((to) => {
  const auth = useAuthStore();
  if (to.matched.some((r) => r.meta.requiresAuth) && !auth.isAuthenticated) {
    return { name: "login" };
  }
  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: "dashboard" };
  }
  // A feature switched off in the profile is not reachable via URL either.
  if (to.meta.feature && auth.preferences && !auth.preferences[to.meta.feature]) {
    return { name: "dashboard" };
  }
  return true;
});

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · RevU` : "RevU";
});
