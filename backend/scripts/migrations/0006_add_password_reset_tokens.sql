BEGIN;

CREATE TABLE IF NOT EXISTS "edv"."password_reset_tokens" (
    "prt_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "prt_usr_id" uuid NOT NULL REFERENCES "edv"."users" ("usr_id"),
    "prt_token_hash" text NOT NULL,
    "prt_expires_at" timestamptz NOT NULL,
    "prt_consumed_at" timestamptz,
    "created_at" timestamptz DEFAULT now(),
    "updated_at" timestamptz DEFAULT now()
);


CREATE UNIQUE INDEX IF NOT EXISTS "idx_password_reset_tokens_token_hash"
    ON "edv"."password_reset_tokens" ("prt_token_hash");

CREATE INDEX IF NOT EXISTS "idx_password_reset_tokens_user"
    ON "edv"."password_reset_tokens" ("prt_usr_id");

-- Speeds up "invalidate all outstanding tokens for this user" on a new request.
CREATE INDEX IF NOT EXISTS "idx_password_reset_tokens_user_unconsumed"
    ON "edv"."password_reset_tokens" ("prt_usr_id")
    WHERE "prt_consumed_at" IS NULL;

COMMIT;
