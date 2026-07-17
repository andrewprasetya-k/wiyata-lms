BEGIN;

ALTER TABLE "edv"."school_registration_requests"
    ADD COLUMN IF NOT EXISTS "srr_usr_id" uuid;

COMMENT ON COLUMN "edv"."school_registration_requests"."srr_usr_id" IS
    'Account that submitted this request (Phase 2+). Nullable for backward compatibility: rows created before this migration have no linked account and are rejected by Approve() rather than silently approved without an owning user.';


DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'school_registration_requests_srr_usr_id_fkey'
    ) THEN
        ALTER TABLE "edv"."school_registration_requests"
            ADD CONSTRAINT "school_registration_requests_srr_usr_id_fkey"
            FOREIGN KEY ("srr_usr_id") REFERENCES "edv"."users" ("usr_id");
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS "idx_school_registration_requests_pending_requester"
    ON "edv"."school_registration_requests" ("srr_usr_id")
    WHERE "srr_status" = 'pending' AND "srr_usr_id" IS NOT NULL;

CREATE INDEX IF NOT EXISTS "idx_school_registration_requests_requester"
    ON "edv"."school_registration_requests" ("srr_usr_id");

COMMIT;
