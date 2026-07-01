import { computed, onMounted, onUnmounted, ref } from "vue";
import { getChatRooms } from "../services/chat";
import { connectChatSocket } from "../services/chatSocket";
import { getActiveSchoolId, getStoredToken } from "../services/session";
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
}

async function handleVisibilityChange() {
  if (document.visibilityState !== "visible") return;
  await refreshRooms();
}

function connectRealtime() {
  socketConnection?.close();
  socketConnection = null;
  if (!getStoredToken() || !getActiveSchoolId()) return;

  socketConnection = connectChatSocket({
    onEvent(event: ChatSocketEvent) {
      if (
        event.type === "new_message" ||
        event.type === "message_read" ||
        event.type === "room_updated"
      ) {
        scheduleRefresh();
      }
    },
    onStatusChange(status) {
      if (!isStarted) return;
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

  loading.value = true;
  error.value = "";
  inFlightRefresh = getChatRooms()
    .then((nextRooms) => {
      if (contextKey === requestContext) {
        rooms.value = nextRooms;
      }
      return nextRooms;
    })
    .catch(() => {
      if (contextKey === requestContext) {
        error.value = "Ringkasan chat belum bisa dimuat.";
      }
      return rooms.value;
    })
    .finally(() => {
      if (contextKey === requestContext) {
        loading.value = false;
      }
      inFlightRefresh = null;
    });

  return inFlightRefresh;
}

function ensureContext() {
  const nextContextKey = buildContextKey();
  if (nextContextKey === contextKey) return contextKey;

  contextKey = nextContextKey;
  rooms.value = [];
  error.value = "";
  inFlightRefresh = null;
  socketConnection?.close();
  socketConnection = null;

  if (isStarted) {
    connectRealtime();
  }

  return contextKey;
}

function buildContextKey() {
  const token = getStoredToken();
  const schoolId = getActiveSchoolId();
  if (!token || !schoolId) return "";
  return `${schoolId}:${token}`;
}
