import { getStoredToken, getActiveSchoolId } from "./session";

export type SidebarStreamEventType = "notification_changed" | "feed_changed";

export interface SidebarStreamEvent {
  type: SidebarStreamEventType;
  schoolId?: string;
  userId?: string;
  payload?: Record<string, unknown>;
}

type SidebarStreamListener = (event: SidebarStreamEvent) => void;

let eventSource: EventSource | null = null;
let subscriberCount = 0;
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

function startSidebarStream() {
  stopSidebarStream();
  const url = buildSidebarStreamUrl();
  if (!url) return;

  eventSource = new EventSource(url);
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
    // EventSource auto-reconnects; keep the connection open.
  };
}

function stopSidebarStream() {
  eventSource?.close();
  eventSource = null;
}

function handleContextChanged() {
  startSidebarStream();
}

function buildSidebarStreamUrl() {
  const token = getStoredToken();
  const schoolId = getActiveSchoolId();
  if (!token || !schoolId) return "";

  const apiBase =
    import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api";
  const url = new URL(
    `${apiBase.replace(/\/$/, "")}/events/sidebar`,
    window.location.origin,
  );
  url.searchParams.set("token", token);
  url.searchParams.set("schoolId", schoolId);
  return url.toString();
}
