/**
 * Shared error categorization for consistent UX across the app.
 * See docs/ERROR_HANDLING.md for the full convention and rationale.
 */
export type ErrorCategory =
  | "validation"
  | "permission"
  | "not_found"
  | "business_rule"
  | "network"
  | "unexpected";

export function classifyApiError(error: unknown): ErrorCategory {
  if (typeof error !== "object" || error === null || !("response" in error)) {
    // Axios throws a plain Error (no `response`) for network failures,
    // aborted requests, and timeouts.
    return "network";
  }

  const status = (
    error as { response?: { status?: number } }
  ).response?.status;

  if (status === 400 || status === 422) return "validation";
  if (status === 401 || status === 403) return "permission";
  if (status === 404) return "not_found";
  if (status === 409) return "business_rule";
  if (status && status >= 500) return "unexpected";
  return "unexpected";
}
