import { computed, onMounted, onUnmounted, ref } from "vue";
import { getFeedUnreadCount } from "../services/feed";
import { getActiveRole, getActiveSchoolId } from "../services/session";
import { getAccessToken } from "../services/accessToken";
import { subscribeSidebarStream } from "../services/sidebarStream";

const unreadCount = ref(0);
let pollTimer: number | undefined;
let refreshTimer: number | undefined;
let consumerCount = 0;
let optimisticClearVersion = 0;
let suppressAutoRefreshUntil = 0;
let requestGeneration = 0;
let unsubscribeSidebarStream: (() => void) | null = null;
const optimisticClearSuppressionMs = 5000;

function isAutoRefreshSuppressed() {
  return Date.now() < suppressAutoRefreshUntil;
}

async function refreshUnreadCount(options: { force?: boolean } = {}) {
  const requestContext = buildContextKey();
  if (!requestContext) {
    unreadCount.value = 0;
    return;
  }

  if (!options.force && isAutoRefreshSuppressed()) {
    return;
  }

  const generation = requestGeneration;
  try {
    const response = await getFeedUnreadCount();
    if (
      generation === requestGeneration &&
      requestContext === buildContextKey()
    ) {
      unreadCount.value = Math.max(0, response.unreadCount || 0);
    }
  } catch {
    // Feed badge should fail silently.
  }
}

function scheduleRefresh() {
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
  }
  refreshTimer = window.setTimeout(() => {
    void refreshUnreadCount();
  }, 150);
}

function handleVisibilityChange() {
  if (document.visibilityState !== "visible") return;
  void refreshUnreadCount();
}

function handleFeedUnreadRefresh() {
  scheduleRefresh();
}

function handleSidebarEvent(event: { type: string }) {
  if (event.type === "notification_changed" || event.type === "feed_changed") {
    scheduleRefresh();
  }
}

function handleContextChanged() {
  requestGeneration += 1;
  suppressAutoRefreshUntil = 0;
  unreadCount.value = 0;
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
    refreshTimer = undefined;
  }
  if (consumerCount > 0) {
    void refreshUnreadCount({ force: true });
  }
}

function buildContextKey() {
  const token = getAccessToken();
  const schoolId = getActiveSchoolId();
  const role = getActiveRole();
  if (!token || !schoolId || !role) return "";
  return `${schoolId}:${role}:${token}`;
}

export function emitFeedUnreadRefresh() {
  window.dispatchEvent(new Event("wiyata:feed-unread-refresh"));
}

export function clearFeedUnreadOptimistically() {
  optimisticClearVersion += 1;
  suppressAutoRefreshUntil = Date.now() + optimisticClearSuppressionMs;
  const snapshot = {
    previousCount: unreadCount.value,
    version: optimisticClearVersion,
  };
  unreadCount.value = 0;
  return snapshot;
}

export function restoreFeedUnreadCount(snapshot: {
  previousCount: number;
  version: number;
}) {
  if (snapshot.version !== optimisticClearVersion) return;
  suppressAutoRefreshUntil = 0;
  unreadCount.value = Math.max(0, snapshot.previousCount);
}

export function useFeedUnreadCount() {
  const badgeLabel = computed(() => {
    if (unreadCount.value <= 0) return "";
    return unreadCount.value > 99 ? "99+" : String(unreadCount.value);
  });

  onMounted(() => {
    consumerCount += 1;
    if (consumerCount === 1) {
      void refreshUnreadCount();
      document.addEventListener("visibilitychange", handleVisibilityChange);
      window.addEventListener(
        "wiyata:feed-unread-refresh",
        handleFeedUnreadRefresh,
      );
      window.addEventListener("wiyata:context-changed", handleContextChanged);
      unsubscribeSidebarStream = subscribeSidebarStream(handleSidebarEvent);
    }
  });

  onUnmounted(() => {
    consumerCount = Math.max(0, consumerCount - 1);
    if (consumerCount === 0) {
      if (pollTimer) {
        window.clearTimeout(pollTimer);
      }
      if (refreshTimer) {
        window.clearTimeout(refreshTimer);
      }
      if (pollTimer) {
        window.clearTimeout(pollTimer);
      }
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      window.removeEventListener(
        "wiyata:feed-unread-refresh",
        handleFeedUnreadRefresh,
      );
      window.removeEventListener(
        "wiyata:context-changed",
        handleContextChanged,
      );
      unsubscribeSidebarStream?.();
      unsubscribeSidebarStream = null;
    }
  });

  return {
    unreadCount,
    badgeLabel,
    refreshUnreadCount,
  };
}
