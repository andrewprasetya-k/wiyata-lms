BEGIN;

-- Phase 11.5.1 (brute-force detection now also groups by source IP, not
-- just target email — see SecurityRepository.GroupFailedLoginsByIP /
-- GetFailedLoginAttemptTimesByIP). That query shape is
-- WHERE log_action = 'auth.login.failed' AND ip_address = ? AND
-- created_at >= ?, grouped/ordered by created_at — idx_logs_action_created_at
-- (migration 0010) covers the action+date filter but would still need a
-- row-by-row filter on ip_address after the index scan. A partial index
-- (ip_address IS NOT NULL) both serves this query directly and, since IP
-- capture only started with this phase, permanently excludes every
-- pre-existing log row that will never have an IP — those rows correctly
-- never match an IP-based lookup, no reason to carry them in this index.
CREATE INDEX IF NOT EXISTS "idx_logs_action_ip_created_at"
    ON "edv"."logs" ("log_action", "ip_address", "created_at" DESC)
    WHERE "ip_address" IS NOT NULL;

COMMIT;
