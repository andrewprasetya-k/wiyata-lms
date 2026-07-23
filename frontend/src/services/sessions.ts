import { api } from "./api";

export interface Session {
  id: string;
  loggedInAt: string;
  expiresAt: string;
  userAgent: string;
  ipAddress: string;
}

export async function listSessions(): Promise<Session[]> {
  const { data } = await api.get<Session[]>("/me/sessions");
  return data;
}

export async function revokeSession(id: string): Promise<void> {
  await api.delete(`/me/sessions/${encodeURIComponent(id)}`);
}
