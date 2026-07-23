import { getActiveSchoolId } from "./session";
import { fetchWsTicket } from "./wsTicket";

export type SidebarStreamEventType = "notification_changed" | "feed_changed";

export interface SidebarStreamEvent {
  type: SidebarStreamEventType;
  schoolId?: string;
  userId?: string;
  payload?: Record<string, unknown>;
}

type SidebarStreamListener = (event: SidebarStreamEvent) => void;

const reconnectDelaysMs = [1000, 2000, 5000, 10000];

let eventSource: EventSource | null = null;
let subscriberCount = 0;
let reconnectTimer: number | undefined;
let retryIndex = 0;
let closedByClient = false;
const listeners = new Set<SidebarStreamListener>();

export function subscribeSidebarStream(listener: SidebarStreamListener) {
  listeners.add(listener);
  subscriberCount += 1;
  if (subscriberCount === 1) {
    startSidebarStream();
    window.addEventListener("wiyata:context-changed", handleContextChanged);
  }

  return () => {
    listeners.delete(listener);
    subscriberCount = Math.max(0, subscriberCount - 1);
    if (subscriberCount === 0) {
      stopSidebarStream();
      window.removeEventListener(
        "wiyata:context-changed",
        handleContextChanged,
      );
    }
  };
}

function disconnectCurrent() {
  eventSource?.close();
  eventSource = null;
}

function stopSidebarStream() {
  closedByClient = true;
  if (reconnectTimer !== undefined) {
    window.clearTimeout(reconnectTimer);
    reconnectTimer = undefined;
  }
  disconnectCurrent();
}

async function startSidebarStream() {
  disconnectCurrent();
  closedByClient = false;

  const schoolId = getActiveSchoolId();
  if (!schoolId) return;

  // Fetched fresh on every (re)connect attempt — a WS ticket is single-use,
  // so reusing one from a prior attempt would never work. This is also why
  // reconnection below can't rely on EventSource's own built-in auto-retry:
  // that always re-requests the exact URL it was constructed with, which
  // would carry an already-consumed ticket and fail forever.
  const url = await buildSidebarStreamUrl(schoolId);
  if (closedByClient || !url) return;

  eventSource = new EventSource(url);
  eventSource.onopen = () => {
    retryIndex = 0;
  };
  eventSource.onmessage = (message) => {
    try {
      const event = JSON.parse(message.data) as SidebarStreamEvent;
      if (!event?.type) return;
      for (const listener of listeners) {
        listener(event);
      }
    } catch {
      // Ignore malformed SSE events and keep REST as source of truth.
    }
  };
  eventSource.onerror = () => {
    disconnectCurrent();
    if (closedByClient || subscriberCount === 0) return;
    const delay =
      reconnectDelaysMs[Math.min(retryIndex, reconnectDelaysMs.length - 1)];
    retryIndex += 1;
    reconnectTimer = window.setTimeout(startSidebarStream, delay);
  };
}

function handleContextChanged() {
  retryIndex = 0;
  startSidebarStream();
}

async function buildSidebarStreamUrl(schoolId: string): Promise<string> {
  const ticket = await fetchWsTicket();
  if (!ticket) return "";

  const apiBase =
    import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api";
  const url = new URL(
    `${apiBase.replace(/\/$/, "")}/events/sidebar`,
    window.location.origin,
  );
  url.searchParams.set("ticket", ticket);
  url.searchParams.set("schoolId", schoolId);
  return url.toString();
}
