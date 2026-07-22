BEGIN;

DROP INDEX IF EXISTS "edv"."idx_logs_severity";

CREATE INDEX IF NOT EXISTS "idx_logs_severity"
    ON "edv"."logs" ("severity");

COMMIT;
