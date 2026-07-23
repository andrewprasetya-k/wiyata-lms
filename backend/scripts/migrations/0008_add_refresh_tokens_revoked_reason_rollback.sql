BEGIN;

ALTER TABLE "edv"."refresh_tokens"
    DROP COLUMN IF EXISTS "rft_revoked_reason";

COMMIT;
