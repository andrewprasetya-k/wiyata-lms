BEGIN;

DROP INDEX IF EXISTS "edv"."idx_email_verifications_user_unconsumed";
DROP INDEX IF EXISTS "edv"."idx_email_verifications_user";
DROP INDEX IF EXISTS "edv"."idx_email_verifications_token_hash";

DROP TABLE IF EXISTS "edv"."email_verifications";

ALTER TABLE "edv"."users"
    DROP COLUMN IF EXISTS "usr_email_verified_at";

COMMIT;
