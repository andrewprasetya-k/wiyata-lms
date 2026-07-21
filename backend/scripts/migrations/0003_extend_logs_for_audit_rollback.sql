BEGIN;

ALTER TABLE "edv"."logs"
    DROP CONSTRAINT IF EXISTS "logs_actor_school_user_id_fkey";

ALTER TABLE "edv"."logs"
    DROP COLUMN IF EXISTS "correlation_id",
    DROP COLUMN IF EXISTS "user_agent",
    DROP COLUMN IF EXISTS "ip_address",
    DROP COLUMN IF EXISTS "severity",
    DROP COLUMN IF EXISTS "scope",
    DROP COLUMN IF EXISTS "entity_id",
    DROP COLUMN IF EXISTS "entity_type",
    DROP COLUMN IF EXISTS "actor_school_user_id";

COMMIT;
