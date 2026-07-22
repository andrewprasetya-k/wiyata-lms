BEGIN;

DROP INDEX IF EXISTS "edv"."idx_logs_severity";
DROP INDEX IF EXISTS "edv"."idx_logs_correlation_id";
DROP INDEX IF EXISTS "edv"."idx_logs_user_created_at";
DROP INDEX IF EXISTS "edv"."idx_logs_school_created_at";

COMMIT;
