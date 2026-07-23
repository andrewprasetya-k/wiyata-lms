// The access token is short-lived (15 min) as of the refresh-token
// migration and deliberately lives only in memory, not localStorage —
// losing it on a hard reload just costs one silent refresh-cookie round
// trip (see stores/auth.ts's restoreSession), which the app already needs
// to support. A plain module-level singleton (rather than reading this
// from the Pinia auth store) avoids api.ts depending on stores/auth.ts,
// which already depends on api.ts — a circular import.
let currentAccessToken: string | null = null;

export function getAccessToken(): string | null {
  return currentAccessToken;
}

export function setAccessToken(token: string | null): void {
  currentAccessToken = token;
}
