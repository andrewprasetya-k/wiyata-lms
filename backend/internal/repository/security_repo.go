package repository

import (
	"backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

// FailedLoginGroup is one row of the "failures grouped by target email"
// aggregate — the pre-filter step of brute-force detection (see
// SecurityRepository.GetFailedLoginAttemptTimes for the follow-up density
// check). Login is pre-authentication, so there is no user_id to group by
// yet — email (from auth.login.failed's metadata) is the only identity
// signal available.
type FailedLoginGroup struct {
	Email         string
	FailureCount  int64
	LastAttemptAt time.Time
}

// FailedLoginIPGroup is GroupFailedLoginsByIP's per-source-IP counterpart
// to FailedLoginGroup. IP capture on auth.login.failed only started with
// Phase 11.5.1 (see AuthService.logLoginFailed) — every query behind this
// type explicitly excludes rows with a NULL ip_address, so pre-existing
// log rows are correctly treated as "IP unknown," never miscounted as a
// shared/matching IP.
type FailedLoginIPGroup struct {
	IPAddress     string
	FailureCount  int64
	LastAttemptAt time.Time
}

// SuspiciousActivityFilter selects the log rows a "suspicious activity"
// widget cares about — a fixed action list plus a time window, optionally
// pinned to one school. See SecurityRepository for how the school pin is
// resolved for actions that carry a user_id vs. actions that don't.
type SecurityRepository interface {
	// CountByActions counts logs matching any of actions within
	// [since, now), scoped to schoolID when non-nil (super_admin passes nil
	// for an unrestricted, platform-wide count).
	CountByActions(actions []string, since time.Time, schoolID *string) (int64, error)

	// GroupFailedLoginsByEmail aggregates auth.login.failed rows in
	// [since, now) by target email, returning only emails whose total
	// failure count already meets minCount — the caller runs a further
	// density check (GetFailedLoginAttemptTimes) on just this shortlist.
	GroupFailedLoginsByEmail(since time.Time, minCount int, schoolID *string) ([]FailedLoginGroup, error)

	// GetFailedLoginAttemptTimes returns every auth.login.failed timestamp
	// for one target email in [since, now), ascending — used to check
	// whether minCount failures fall inside any window-length span.
	GetFailedLoginAttemptTimes(email string, since time.Time) ([]time.Time, error)

	// GroupFailedLoginsByIP is GroupFailedLoginsByEmail's per-source-IP
	// counterpart — same shortlist-then-density-check shape, but grouped by
	// the caller's IP instead of the target account. Rows with no IP
	// (logged before Phase 11.5.1, or a request where ClientIP() came back
	// empty) are excluded entirely, never grouped together as if they
	// shared an IP.
	GroupFailedLoginsByIP(since time.Time, minCount int, schoolID *string) ([]FailedLoginIPGroup, error)

	// GetFailedLoginAttemptTimesByIP is GetFailedLoginAttemptTimes's
	// per-IP counterpart.
	GetFailedLoginAttemptTimesByIP(ipAddress string, since time.Time) ([]time.Time, error)

	// GetRecentByActions returns the most recent logs matching any of
	// actions within [since, now), newest first, scoped to schoolID when
	// non-nil.
	GetRecentByActions(actions []string, since time.Time, schoolID *string, limit int) ([]*domain.Log, error)
}

type securityRepository struct {
	db *gorm.DB
}

func NewSecurityRepository(db *gorm.DB) SecurityRepository {
	return &securityRepository{db: db}
}

// scopeToSchool restricts a logs-table query to rows whose actor (or, for
// pre-auth rows with no user_id, whose metadata email) belongs to schoolID.
// Every action this dashboard reads is scope=platform with no log_sch_id
// (see log.md — auth.* actions never carry a school_id), so a school pin
// can't be a simple `WHERE log_sch_id = ?` like the rest of the audit log
// surface; it has to resolve membership indirectly instead.
func scopeToSchool(query *gorm.DB, schoolID *string) *gorm.DB {
	if schoolID == nil {
		return query
	}
	return query.Where(
		`(
			EXISTS (
				SELECT 1 FROM edv.school_users su
				WHERE su.scu_usr_id = logs.log_usr_id
				AND su.scu_sch_id = ?
				AND su.deleted_at IS NULL
			)
			OR EXISTS (
				SELECT 1 FROM edv.users u
				JOIN edv.school_users su ON su.scu_usr_id = u.usr_id
					AND su.scu_sch_id = ? AND su.deleted_at IS NULL
				WHERE u.usr_email = (logs.log_metadata->>'email')
			)
		)`,
		*schoolID, *schoolID,
	)
}

func (r *securityRepository) CountByActions(actions []string, since time.Time, schoolID *string) (int64, error) {
	var count int64
	query := r.db.Table("edv.logs AS logs").
		Where("logs.log_action IN ?", actions).
		Where("logs.created_at >= ?", since)
	query = scopeToSchool(query, schoolID)
	err := query.Count(&count).Error
	return count, err
}

func (r *securityRepository) GroupFailedLoginsByEmail(since time.Time, minCount int, schoolID *string) ([]FailedLoginGroup, error) {
	var rows []FailedLoginGroup
	query := r.db.Table("edv.logs AS logs").
		Select("logs.log_metadata->>'email' AS email, COUNT(*) AS failure_count, MAX(logs.created_at) AS last_attempt_at").
		Where("logs.log_action = ?", "auth.login.failed").
		Where("logs.created_at >= ?", since).
		Where("logs.log_metadata->>'email' IS NOT NULL")
	query = scopeToSchool(query, schoolID)
	err := query.
		Group("logs.log_metadata->>'email'").
		Having("COUNT(*) >= ?", minCount).
		Order("failure_count DESC").
		Limit(50).
		Scan(&rows).Error
	return rows, err
}

func (r *securityRepository) GetFailedLoginAttemptTimes(email string, since time.Time) ([]time.Time, error) {
	var times []time.Time
	err := r.db.Table("edv.logs AS logs").
		Select("logs.created_at").
		Where("logs.log_action = ?", "auth.login.failed").
		Where("logs.created_at >= ?", since).
		Where("logs.log_metadata->>'email' = ?", email).
		Order("logs.created_at ASC").
		Limit(200).
		Scan(&times).Error
	return times, err
}

func (r *securityRepository) GroupFailedLoginsByIP(since time.Time, minCount int, schoolID *string) ([]FailedLoginIPGroup, error) {
	var rows []FailedLoginIPGroup
	query := r.db.Table("edv.logs AS logs").
		Select("logs.ip_address AS ip_address, COUNT(*) AS failure_count, MAX(logs.created_at) AS last_attempt_at").
		Where("logs.log_action = ?", "auth.login.failed").
		Where("logs.created_at >= ?", since).
		Where("logs.ip_address IS NOT NULL")
	// School scope still resolves via the target account's membership
	// (same scopeToSchool as the email grouping) — this reads as "which
	// IPs attacked accounts belonging to this school," not "which IPs
	// belong to this school" (an IP has no school of its own).
	query = scopeToSchool(query, schoolID)
	err := query.
		Group("logs.ip_address").
		Having("COUNT(*) >= ?", minCount).
		Order("failure_count DESC").
		Limit(50).
		Scan(&rows).Error
	return rows, err
}

func (r *securityRepository) GetFailedLoginAttemptTimesByIP(ipAddress string, since time.Time) ([]time.Time, error) {
	var times []time.Time
	err := r.db.Table("edv.logs AS logs").
		Select("logs.created_at").
		Where("logs.log_action = ?", "auth.login.failed").
		Where("logs.created_at >= ?", since).
		Where("logs.ip_address = ?", ipAddress).
		Order("logs.created_at ASC").
		Limit(200).
		Scan(&times).Error
	return times, err
}

func (r *securityRepository) GetRecentByActions(actions []string, since time.Time, schoolID *string, limit int) ([]*domain.Log, error) {
	var logs []*domain.Log
	query := r.db.Model(&domain.Log{}).Preload("User").
		Where("logs.log_action IN ?", actions).
		Where("logs.created_at >= ?", since)
	query = scopeToSchool(query, schoolID)
	err := query.Order("logs.created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
