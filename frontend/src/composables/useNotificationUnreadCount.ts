import { computed, ref } from "vue";
import { getNotificationUnreadCount } from "../services/notifications";
import { getActiveRole, getActiveSchoolId } from "../services/session";
import { getAccessToken } from "../services/accessToken";
import { subscribeSidebarStream } from "../services/sidebarStream";

const unreadCount = ref(0);
const loading = ref(false);
const error = ref("");
let initialized = false;
let lifecycleStarted = false;
let refreshPromise: Promise<void> | null = null;
let requestGeneration = 0;

function normalizeCount(value: number) {
  return Math.max(0, Number.isFinite(value) ? value : 0);
}

async function refreshUnreadCount() {
  const requestContext = buildContextKey();
  if (!requestContext) {
    unreadCount.value = 0;
    error.value = "";
    return;
  }

  if (refreshPromise) return refreshPromise;

  const generation = requestGeneration;
  loading.value = true;
  error.value = "";

  const request = getNotificationUnreadCount()
    .then((response) => {
      if (
        generation === requestGeneration &&
        requestContext === buildContextKey()
      ) {
        unreadCount.value = normalizeCount(response.unreadCount);
      }
    })
    .catch(() => {
      if (
        generation === requestGeneration &&
        requestContext === buildContextKey()
      ) {
        error.value = "Jumlah notifikasi belum bisa dimuat.";
      }
    })
    .finally(() => {
      if (
        generation === requestGeneration &&
        requestContext === buildContextKey()
      ) {
        loading.value = false;
      }
      if (refreshPromise === request) {
        refreshPromise = null;
      }
    });

  refreshPromise = request;
  return refreshPromise;
}

function startLifecycleRefresh() {
  if (lifecycleStarted || typeof document === "undefined") return;
  lifecycleStarted = true;

  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") {
      void refreshUnreadCount();
    }
  });
  window.addEventListener("wiyata:context-changed", handleContextChanged);
  subscribeSidebarStream((event) => {
    if (
      event.type === "notification_changed" ||
      event.type === "feed_changed"
    ) {
      void refreshUnreadCount();
    }
  });
}

function handleContextChanged() {
  requestGeneration += 1;
  initialized = false;
  refreshPromise = null;
  unreadCount.value = 0;
  loading.value = false;
  error.value = "";
  if (lifecycleStarted) {
    initialized = true;
    void refreshUnreadCount();
  }
}

function buildContextKey() {
  const token = getAccessToken();
  const schoolId = getActiveSchoolId();
  const role = getActiveRole();
  if (!token || !schoolId || !role) return "";
  return `${schoolId}:${role}:${token}`;
}

export function useNotificationUnreadCount() {
  startLifecycleRefresh();

  if (!initialized) {
    initialized = true;
    void refreshUnreadCount();
  }

  function set(value: number) {
    unreadCount.value = normalizeCount(value);
  }

  function decrement(step = 1) {
    unreadCount.value = normalizeCount(unreadCount.value - step);
  }

  function clear() {
    unreadCount.value = 0;
  }

  const badgeLabel = computed(() => {
    if (unreadCount.value <= 0) return "";
    return unreadCount.value > 99 ? "99+" : String(unreadCount.value);
  });

  return {
    unreadCount,
    loading,
    error,
    badgeLabel,
    refresh: refreshUnreadCount,
    set,
    decrement,
    clear,
  };
}
