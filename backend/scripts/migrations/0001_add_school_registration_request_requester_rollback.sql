BEGIN;

DROP INDEX IF EXISTS "edv"."idx_school_registration_requests_requester";
DROP INDEX IF EXISTS "edv"."idx_school_registration_requests_pending_requester";

ALTER TABLE "edv"."school_registration_requests"
    DROP CONSTRAINT IF EXISTS "school_registration_requests_srr_usr_id_fkey";

ALTER TABLE "edv"."school_registration_requests"
    DROP COLUMN IF EXISTS "srr_usr_id";

COMMIT;
