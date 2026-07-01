import { computed, onMounted, onUnmounted, ref } from "vue";
import { getFeedUnreadCount } from "../services/feed";

const unreadCount = ref(0);
let pollTimer: number | undefined;
let refreshTimer: number | undefined;
let consumerCount = 0;
let optimisticClearVersion = 0;
let suppressAutoRefreshUntil = 0;
const optimisticClearSuppressionMs = 5000;

function isAutoRefreshSuppressed() {
  return Date.now() < suppressAutoRefreshUntil;
}

function schedulePolling() {
  if (pollTimer) {
    window.clearTimeout(pollTimer);
  }
  pollTimer = window.setTimeout(async () => {
    await refreshUnreadCount();
    if (consumerCount > 0) {
      schedulePolling();
    }
  }, 45000);
}

async function refreshUnreadCount(options: { force?: boolean } = {}) {
  if (!options.force && isAutoRefreshSuppressed()) {
    return;
  }

  try {
    const response = await getFeedUnreadCount();
    unreadCount.value = Math.max(0, response.unreadCount || 0);
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

export function restoreFeedUnreadCount(snapshot: { previousCount: number; version: number }) {
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
      schedulePolling();
      document.addEventListener("visibilitychange", handleVisibilityChange);
      window.addEventListener("wiyata:feed-unread-refresh", handleFeedUnreadRefresh);
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
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      window.removeEventListener("wiyata:feed-unread-refresh", handleFeedUnreadRefresh);
    }
  });

  return {
    unreadCount,
    badgeLabel,
    refreshUnreadCount,
  };
}
