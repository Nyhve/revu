import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import { router } from "./router";
import { useAuthStore } from "./stores/auth";
import { setUnauthorizedHandler } from "./api/client";
import "./index.css";

const app = createApp(App);
app.use(createPinia());
app.use(router);

// An expired or invalid token logs the user out instead of showing errors forever.
const auth = useAuthStore();
setUnauthorizedHandler(() => {
  auth.logout();
  if (router.currentRoute.value.meta.requiresAuth) {
    router.replace({ name: "login" });
  }
});
if (auth.isAuthenticated) {
  void auth.loadPreferences();
}

app.mount("#app");
