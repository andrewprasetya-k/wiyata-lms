import { useAuthStore } from "../stores/auth";

let lifecycleStarted = false;
let lastRefreshAt = 0;
const REFRESH_COOLDOWN_MS = 3000;

function refreshIfNeeded() {
  const auth = useAuthStore();
  if (!auth.isAuthenticated) return;

  const now = Date.now();
  if (now - lastRefreshAt < REFRESH_COOLDOWN_MS) return;
  lastRefreshAt = now;

  void auth.refreshUserContext();
}

function handleVisibilityChange() {
  if (document.visibilityState !== "visible") return;
  refreshIfNeeded();
}

function handleWindowFocus() {
  refreshIfNeeded();
}

function startLifecycleRefresh() {
  if (lifecycleStarted || typeof document === "undefined") return;
  lifecycleStarted = true;

  document.addEventListener("visibilitychange", handleVisibilityChange);
  window.addEventListener("focus", handleWindowFocus);
}

export function useAuthContextSync() {
  startLifecycleRefresh();
}
