BEGIN;

CREATE INDEX IF NOT EXISTS "idx_logs_school_created_at"
    ON "edv"."logs" ("log_sch_id", "created_at" DESC);

CREATE INDEX IF NOT EXISTS "idx_logs_user_created_at"
    ON "edv"."logs" ("log_usr_id", "created_at" DESC);
CREATE INDEX IF NOT EXISTS "idx_logs_correlation_id"
    ON "edv"."logs" ("correlation_id");

CREATE INDEX IF NOT EXISTS "idx_logs_severity"
    ON "edv"."logs" ("severity");

COMMIT;
