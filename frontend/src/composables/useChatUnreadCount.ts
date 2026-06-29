import { computed, onMounted, onUnmounted, ref } from "vue";
import { getChatRooms } from "../services/chat";
import { connectChatSocket } from "../services/chatSocket";
import type { ChatSocketEvent } from "../types/chat";

export function useChatUnreadCount() {
  const unreadCount = ref(0);
  const isConnected = ref(false);
  let pollTimer: number | undefined;
  let refreshTimer: number | undefined;
  let socketConnection: { close: () => void } | null = null;
  let isDestroyed = false;

  const badgeLabel = computed(() => {
    if (unreadCount.value <= 0) return "";
    return unreadCount.value > 99 ? "99+" : String(unreadCount.value);
  });

  onMounted(() => {
    isDestroyed = false;
    void refreshUnreadCount();
    connectRealtime();
    scheduleNextPoll();
    document.addEventListener("visibilitychange", handleVisibilityChange);
  });

  onUnmounted(() => {
    isDestroyed = true;
    if (pollTimer) {
      window.clearTimeout(pollTimer);
    }
    if (refreshTimer) {
      window.clearTimeout(refreshTimer);
    }
    socketConnection?.close();
    document.removeEventListener("visibilitychange", handleVisibilityChange);
  });

  async function handleVisibilityChange() {
    if (document.visibilityState !== "visible") return;
    await refreshUnreadCount();
  }

  function connectRealtime() {
    socketConnection?.close();
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
        isConnected.value = status === "connected";
        scheduleNextPoll();
      },
    });
  }

  function scheduleRefresh() {
    if (refreshTimer) {
      window.clearTimeout(refreshTimer);
    }
    refreshTimer = window.setTimeout(() => {
      void refreshUnreadCount();
    }, 150);
  }

  function scheduleNextPoll() {
    if (isDestroyed) return;
    if (pollTimer) {
      window.clearTimeout(pollTimer);
    }
    const delay = isConnected.value ? 90000 : 18000;
    pollTimer = window.setTimeout(async () => {
      await refreshUnreadCount();
      scheduleNextPoll();
    }, delay);
  }

  async function refreshUnreadCount() {
    try {
      const rooms = await getChatRooms();
      unreadCount.value = rooms.reduce(
        (total, room) => total + Math.max(0, room.unreadCount || 0),
        0,
      );
    } catch {
      // Sidebar unread badge should fail silently.
    }
  }

  return {
    unreadCount,
    badgeLabel,
    refreshUnreadCount,
  };
}
