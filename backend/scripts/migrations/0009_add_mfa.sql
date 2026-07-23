BEGIN;

ALTER TABLE "edv"."users"
    ADD COLUMN IF NOT EXISTS "usr_mfa_grace_started_at" timestamptz;


CREATE TABLE IF NOT EXISTS "edv"."user_mfa" (
    "umf_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "umf_usr_id" uuid NOT NULL REFERENCES "edv"."users" ("usr_id"),
    "umf_secret_encrypted" text NOT NULL,
    "umf_enabled_at" timestamptz,
    "created_at" timestamptz DEFAULT now(),
    "updated_at" timestamptz DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_user_mfa_user"
    ON "edv"."user_mfa" ("umf_usr_id");


CREATE TABLE IF NOT EXISTS "edv"."mfa_recovery_codes" (
    "mrc_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "mrc_usr_id" uuid NOT NULL REFERENCES "edv"."users" ("usr_id"),
    "mrc_code_hash" text NOT NULL,
    "mrc_consumed_at" timestamptz,
    "created_at" timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS "idx_mfa_recovery_codes_user"
    ON "edv"."mfa_recovery_codes" ("mrc_usr_id");

CREATE UNIQUE INDEX IF NOT EXISTS "idx_mfa_recovery_codes_hash"
    ON "edv"."mfa_recovery_codes" ("mrc_code_hash");


-- Pre-auth token: a short-lived (~10 min), single-use token proving "this
-- device already presented the correct password for this user", issued by
-- AuthService.Login instead of a real access/refresh token pair whenever a
-- second step (MFA code, or forced enrollment) is still required. Chosen as
-- a DB table (not an in-memory store like the WS ticket) because this sits
-- on the core authentication path — durability across a server
-- restart/horizontal scaling matters more here than it did for a
-- WS-handshake convenience ticket, and the existing token-table shape
-- (email_verifications/password_reset_tokens) already fits perfectly.
CREATE TABLE IF NOT EXISTS "edv"."mfa_preauth_tokens" (
    "mpt_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "mpt_usr_id" uuid NOT NULL REFERENCES "edv"."users" ("usr_id"),
    "mpt_token_hash" text NOT NULL,
    "mpt_purpose" text NOT NULL,
    "mpt_expires_at" timestamptz NOT NULL,
    "mpt_consumed_at" timestamptz,
    "created_at" timestamptz DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_mfa_preauth_tokens_hash"
    ON "edv"."mfa_preauth_tokens" ("mpt_token_hash");

CREATE INDEX IF NOT EXISTS "idx_mfa_preauth_tokens_user"
    ON "edv"."mfa_preauth_tokens" ("mpt_usr_id");

COMMIT;
