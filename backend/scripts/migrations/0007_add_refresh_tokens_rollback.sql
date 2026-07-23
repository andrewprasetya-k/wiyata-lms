BEGIN;

DROP INDEX IF EXISTS "edv"."idx_refresh_tokens_user";
DROP INDEX IF EXISTS "edv"."idx_refresh_tokens_family_id";
DROP INDEX IF EXISTS "edv"."idx_refresh_tokens_token_hash";

DROP TABLE IF EXISTS "edv"."refresh_tokens";

COMMIT;
