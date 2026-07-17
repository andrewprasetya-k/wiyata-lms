BEGIN;

ALTER TABLE "edv"."users"
    ADD COLUMN IF NOT EXISTS "usr_email_verified_at" timestamptz;


CREATE TABLE IF NOT EXISTS "edv"."email_verifications" (
    "evf_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "evf_usr_id" uuid NOT NULL REFERENCES "edv"."users" ("usr_id"),
    "evf_token_hash" text NOT NULL,
    "evf_expires_at" timestamptz NOT NULL,
    "evf_consumed_at" timestamptz,
    "created_at" timestamptz DEFAULT now(),
    "updated_at" timestamptz DEFAULT now()
);


CREATE UNIQUE INDEX IF NOT EXISTS "idx_email_verifications_token_hash"
    ON "edv"."email_verifications" ("evf_token_hash");

CREATE INDEX IF NOT EXISTS "idx_email_verifications_user"
    ON "edv"."email_verifications" ("evf_usr_id");

-- Speeds up "invalidate all outstanding tokens for this user" on resend/verify.
CREATE INDEX IF NOT EXISTS "idx_email_verifications_user_unconsumed"
    ON "edv"."email_verifications" ("evf_usr_id")
    WHERE "evf_consumed_at" IS NULL;

COMMIT;
