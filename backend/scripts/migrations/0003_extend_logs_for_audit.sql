BEGIN;

ALTER TABLE "edv"."logs"
    ADD COLUMN IF NOT EXISTS "actor_school_user_id" uuid,
    ADD COLUMN IF NOT EXISTS "entity_type" text,
    ADD COLUMN IF NOT EXISTS "entity_id" uuid,
    ADD COLUMN IF NOT EXISTS "scope" text,
    ADD COLUMN IF NOT EXISTS "severity" text,
    ADD COLUMN IF NOT EXISTS "ip_address" text,
    ADD COLUMN IF NOT EXISTS "user_agent" text,
    ADD COLUMN IF NOT EXISTS "correlation_id" uuid;


DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'logs_actor_school_user_id_fkey'
    ) THEN
        ALTER TABLE "edv"."logs"
            ADD CONSTRAINT "logs_actor_school_user_id_fkey"
            FOREIGN KEY ("actor_school_user_id") REFERENCES "edv"."school_users" ("scu_id");
    END IF;
END $$;

COMMIT;
