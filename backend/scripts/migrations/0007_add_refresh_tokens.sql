BEGIN;

CREATE TABLE IF NOT EXISTS "edv"."refresh_tokens" (
    "rft_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "rft_usr_id" uuid NOT NULL REFERENCES "edv"."users" ("usr_id"),
    "rft_token_hash" text NOT NULL,
    "rft_family_id" uuid NOT NULL,
    "rft_expires_at" timestamptz NOT NULL,
    "rft_revoked_at" timestamptz,
    "rft_user_agent" text,
    "rft_ip_address" text,
    "created_at" timestamptz DEFAULT now(),
    "updated_at" timestamptz DEFAULT now()
);


CREATE UNIQUE INDEX IF NOT EXISTS "idx_refresh_tokens_token_hash"
    ON "edv"."refresh_tokens" ("rft_token_hash");

CREATE INDEX IF NOT EXISTS "idx_refresh_tokens_family_id"
    ON "edv"."refresh_tokens" ("rft_family_id");

CREATE INDEX IF NOT EXISTS "idx_refresh_tokens_user"
    ON "edv"."refresh_tokens" ("rft_usr_id");

COMMIT;
