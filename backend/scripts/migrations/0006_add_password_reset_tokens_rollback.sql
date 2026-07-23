BEGIN;

DROP INDEX IF EXISTS "edv"."idx_password_reset_tokens_user_unconsumed";
DROP INDEX IF EXISTS "edv"."idx_password_reset_tokens_user";
DROP INDEX IF EXISTS "edv"."idx_password_reset_tokens_token_hash";

DROP TABLE IF EXISTS "edv"."password_reset_tokens";

COMMIT;
