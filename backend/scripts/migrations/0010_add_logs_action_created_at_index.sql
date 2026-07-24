BEGIN;

-- Phase 11.5 (Security Dashboard). Every dashboard widget queries by a
-- fixed set of `log_action` values plus a `created_at` range (e.g.
-- WHERE log_action IN ('auth.login.failed', ...) AND created_at >= X
-- ORDER BY created_at DESC) — a shape none of the existing composite
-- indexes serve well: idx_logs_school_created_at/idx_logs_user_created_at
-- are keyed by school/user (these auth events carry neither — see
-- log.md, all auth.* actions are scope=platform with no school_id, and
-- pre-authentication actions like auth.login.failed have no user_id
-- either), and idx_logs_severity groups many unrelated actions under the
-- same severity tier. A composite (log_action, created_at DESC) index
-- lets one index scan satisfy both the action filter and the sort.
CREATE INDEX IF NOT EXISTS "idx_logs_action_created_at"
    ON "edv"."logs" ("log_action", "created_at" DESC);

COMMIT;
