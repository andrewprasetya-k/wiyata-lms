BEGIN;

ALTER TABLE "edv"."refresh_tokens"
    ADD COLUMN IF NOT EXISTS "rft_revoked_reason" text;

COMMIT;
