BEGIN;

-- idx_logs_severity (Phase 10.15, migration 0004) was a single-column
-- index on severity. The actual query shape it serves — the
-- unrestricted, platform-wide GET /logs?severity=X — always also sorts
-- ORDER BY created_at DESC for pagination, so a single-column index
-- still requires a separate sort step after the filter. Replacing it
-- with a composite (severity, created_at DESC) index lets one index
-- scan satisfy both the filter and the sort/pagination order.
DROP INDEX IF EXISTS "edv"."idx_logs_severity";

CREATE INDEX IF NOT EXISTS "idx_logs_severity"
    ON "edv"."logs" ("severity", "created_at" DESC);

COMMIT;
