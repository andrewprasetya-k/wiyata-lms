import { api } from "./api";

interface WSTicketResponse {
  ticket: string;
}

export async function fetchWsTicket(): Promise<string | null> {
  try {
    const { data } = await api.get<WSTicketResponse>("/me/ws-ticket");
    return data.ticket || null;
  } catch {
    return null;
  }
}
