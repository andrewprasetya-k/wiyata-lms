import { computed, onMounted, onUnmounted, ref } from "vue";
import { getChatRooms } from "../services/chat";
import { connectChatSocket } from "../services/chatSocket";
import { getActiveRole, getActiveSchoolId, getStoredToken } from "../services/session";
import type { ChatRoom, ChatSocketEvent } from "../types/chat";

const rooms = ref<ChatRoom[]>([]);
const loading = ref(false);
const error = ref("");
const isConnected = ref(false);

let subscriberCount = 0;
let isStarted = false;
let pollTimer: number | undefined;
let refreshTimer: number | undefined;
let socketConnection: { close: () => void } | null = null;
let inFlightRefresh: Promise<ChatRoom[]> | null = null;
let contextKey = "";
let contextGeneration = 0;

const totalUnreadCount = computed(() =>
  rooms.value.reduce(
    (total, room) => total + Math.max(0, room.unreadCount || 0),
    0,
  ),
);

const badgeLabel = computed(() => {
  if (totalUnreadCount.value <= 0) return "";
  return totalUnreadCount.value > 99 ? "99+" : String(totalUnreadCount.value);
});

export function useChatRoomSummary() {
  onMounted(() => {
    subscriberCount += 1;
    if (subscriberCount === 1) {
      startSummarySync();
    }
  });

  onUnmounted(() => {
    subscriberCount = Math.max(0, subscriberCount - 1);
    if (subscriberCount === 0) {
      stopSummarySync();
    }
  });

  return {
    rooms,
    totalUnreadCount,
    badgeLabel,
    loading,
    error,
    refreshRooms,
  };
}

function startSummarySync() {
  if (isStarted) return;
  isStarted = true;
  ensureContext();
  void refreshRooms();
  connectRealtime();
  scheduleNextPoll();
  document.addEventListener("visibilitychange", handleVisibilityChange);
  window.addEventListener("wiyata:context-changed", handleContextChanged);
}

function stopSummarySync() {
  isStarted = false;
  if (pollTimer) {
    window.clearTimeout(pollTimer);
    pollTimer = undefined;
  }
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
    refreshTimer = undefined;
  }
  socketConnection?.close();
  socketConnection = null;
  inFlightRefresh = null;
  document.removeEventListener("visibilitychange", handleVisibilityChange);
  window.removeEventListener("wiyata:context-changed", handleContextChanged);
}

async function handleVisibilityChange() {
  if (document.visibilityState !== "visible") return;
  await refreshRooms();
}

function connectRealtime() {
  socketConnection?.close();
  socketConnection = null;
  const socketContext = buildContextKey();
  if (!socketContext) return;

  socketConnection = connectChatSocket({
    onEvent(event: ChatSocketEvent) {
      if (contextKey !== socketContext) return;
      if (
        event.type === "new_message" ||
        event.type === "message_read" ||
        event.type === "room_updated"
      ) {
        scheduleRefresh();
      }
    },
    onStatusChange(status) {
      if (!isStarted || contextKey !== socketContext) return;
      isConnected.value = status === "connected";
      scheduleNextPoll();
    },
  });
}

function scheduleRefresh() {
  if (!isStarted) return;
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
  }
  refreshTimer = window.setTimeout(() => {
    void refreshRooms();
  }, 150);
}

function scheduleNextPoll() {
  if (!isStarted) return;
  if (pollTimer) {
    window.clearTimeout(pollTimer);
  }
  const delay = isConnected.value ? 90000 : 18000;
  pollTimer = window.setTimeout(async () => {
    await refreshRooms();
    scheduleNextPoll();
  }, delay);
}

async function refreshRooms() {
  const requestContext = ensureContext();
  if (!requestContext) {
    rooms.value = [];
    error.value = "";
    return rooms.value;
  }

  if (inFlightRefresh) return inFlightRefresh;

  const generation = contextGeneration;
  loading.value = true;
  error.value = "";
  const request = getChatRooms()
    .then((nextRooms) => {
      if (contextKey === requestContext && generation === contextGeneration) {
        rooms.value = nextRooms;
      }
      return nextRooms;
    })
    .catch(() => {
      if (contextKey === requestContext && generation === contextGeneration) {
        error.value = "Ringkasan chat belum bisa dimuat.";
      }
      return rooms.value;
    })
    .finally(() => {
      if (contextKey === requestContext && generation === contextGeneration) {
        loading.value = false;
      }
      if (inFlightRefresh === request) {
        inFlightRefresh = null;
      }
    });

  inFlightRefresh = request;
  return inFlightRefresh;
}

function ensureContext() {
  const nextContextKey = buildContextKey();
  if (nextContextKey === contextKey) return contextKey;

  contextKey = nextContextKey;
  contextGeneration += 1;
  rooms.value = [];
  error.value = "";
  loading.value = false;
  isConnected.value = false;
  inFlightRefresh = null;
  socketConnection?.close();
  socketConnection = null;

  if (isStarted) {
    connectRealtime();
  }

  return contextKey;
}

function handleContextChanged() {
  contextGeneration += 1;
  contextKey = buildContextKey();
  rooms.value = [];
  error.value = "";
  loading.value = false;
  isConnected.value = false;
  inFlightRefresh = null;
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
    refreshTimer = undefined;
  }
  if (pollTimer) {
    window.clearTimeout(pollTimer);
    pollTimer = undefined;
  }
  socketConnection?.close();
  socketConnection = null;

  if (!isStarted) return;
  connectRealtime();
  void refreshRooms();
  scheduleNextPoll();
}

function buildContextKey() {
  const token = getStoredToken();
  const schoolId = getActiveSchoolId();
  const role = getActiveRole();
  if (!token || !schoolId || !role) return "";
  return `${schoolId}:${role}:${token}`;
}
