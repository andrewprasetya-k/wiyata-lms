BEGIN;

DROP INDEX IF EXISTS "edv"."idx_mfa_preauth_tokens_user";
DROP INDEX IF EXISTS "edv"."idx_mfa_preauth_tokens_hash";
DROP TABLE IF EXISTS "edv"."mfa_preauth_tokens";

DROP INDEX IF EXISTS "edv"."idx_mfa_recovery_codes_hash";
DROP INDEX IF EXISTS "edv"."idx_mfa_recovery_codes_user";
DROP TABLE IF EXISTS "edv"."mfa_recovery_codes";

DROP INDEX IF EXISTS "edv"."idx_user_mfa_user";
DROP TABLE IF EXISTS "edv"."user_mfa";

ALTER TABLE "edv"."users"
    DROP COLUMN IF EXISTS "usr_mfa_grace_started_at";

COMMIT;
