import { fetchWsTicket } from "./wsTicket";
import type { AuditLogEvent } from "../types/auditLog";

type AuditSocketOptions = {
  channel: string;
  onEvent: (event: AuditLogEvent) => void;
  onStatusChange?: (status: AuditSocketStatus) => void;
};

type AuditSocketConnection = {
  close: () => void;
};

export type AuditSocketStatus = "connecting" | "connected" | "disconnected" | "failed";

const reconnectDelaysMs = [1000, 2000, 5000, 10000];

export function connectAuditSocket(
  options: AuditSocketOptions,
): AuditSocketConnection {
  let socket: WebSocket | null = null;
  let reconnectTimer: number | undefined;
  let closedByClient = false;
  let retryIndex = 0;
  let hasOpenedCurrentSocket = false;
  let failedBeforeOpenCount = 0;

  async function connect() {
    setStatus("connecting");
    hasOpenedCurrentSocket = false;

    // Fetched fresh on every (re)connect attempt — a WS ticket is
    // single-use, so reusing one from a prior attempt would never work.
    const url = await buildAuditSocketUrl(options.channel);
    if (closedByClient) return; // caller closed while the ticket fetch was in flight
    if (!url) {
      setStatus("disconnected");
      return;
    }

    socket = new WebSocket(url);
    socket.onopen = () => {
      retryIndex = 0;
      failedBeforeOpenCount = 0;
      hasOpenedCurrentSocket = true;
      setStatus("connected");
    };
    socket.onmessage = (message) => {
      try {
        options.onEvent(JSON.parse(message.data) as AuditLogEvent);
      } catch {
        // Ignore malformed realtime events — REST stays the source of truth.
      }
    };
    socket.onclose = () => {
      socket = null;
      if (!hasOpenedCurrentSocket) {
        failedBeforeOpenCount += 1;
      }
      if (closedByClient) {
        setStatus("disconnected");
        return;
      }
      if (failedBeforeOpenCount >= 5) {
        setStatus("failed");
        return;
      }
      setStatus("disconnected");
      const delay =
        reconnectDelaysMs[Math.min(retryIndex, reconnectDelaysMs.length - 1)];
      retryIndex += 1;
      reconnectTimer = window.setTimeout(connect, delay);
    };
    socket.onerror = () => {
      socket?.close();
    };
  }

  connect();

  return {
    close() {
      closedByClient = true;
      if (reconnectTimer) {
        window.clearTimeout(reconnectTimer);
      }
      socket?.close();
      socket = null;
      setStatus("disconnected");
    },
  };

  function setStatus(status: AuditSocketStatus) {
    options.onStatusChange?.(status);
  }
}

async function buildAuditSocketUrl(channel: string): Promise<string> {
  if (!channel) return "";
  const ticket = await fetchWsTicket();
  if (!ticket) return "";

  const apiBase = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api";
  const url = new URL(`${apiBase.replace(/\/$/, "")}/ws/audit`, window.location.origin);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.searchParams.set("ticket", ticket);
  url.searchParams.set("channel", channel);
  return url.toString();
}
