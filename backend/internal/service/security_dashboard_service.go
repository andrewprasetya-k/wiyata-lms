package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"time"
)

const (
	securityDashboardWindowHours = 24
	// bruteForceMinFailures/bruteForceWindow mirror the rate-limit tiers
	// already established elsewhere in this codebase (mfa_verify and
	// change-password both use 5 failed attempts / 15 minutes) rather than
	// inventing a new threshold — see log.md's auth.mfa.verify.failed row.
	bruteForceMinFailures   = 5
	bruteForceWindow        = 15 * time.Minute
	suspiciousActivityLimit = 50
)

// suspiciousActivityActions is deliberately just these three — see log.md
// §4: auth.token.reuse_detected (HIGH, a replayed/stolen refresh token),
// auth.mfa.recovery_code.used (HIGH, authenticator app access was lost),
// and auth.mfa.verify.failed (MEDIUM, a wrong TOTP/recovery code) are the
// only actions in the taxonomy that represent a credential/session
// integrity concern rather than routine account activity.
var suspiciousActivityActions = []string{
	"auth.token.reuse_detected",
	"auth.mfa.recovery_code.used",
	"auth.mfa.verify.failed",
}

type SecurityDashboardService interface {
	// GetDashboard builds the summary for the given scope — pass nil for
	// an unrestricted, platform-wide view (super_admin only; see the
	// handler for the permission gate), or a school ID to pin every widget
	// to that school's members.
	GetDashboard(schoolID *string) (*dto.SecurityDashboardDTO, error)
}

type securityDashboardService struct {
	repo repository.SecurityRepository
}

func NewSecurityDashboardService(repo repository.SecurityRepository) SecurityDashboardService {
	return &securityDashboardService{repo: repo}
}

func (s *securityDashboardService) GetDashboard(schoolID *string) (*dto.SecurityDashboardDTO, error) {
	since := time.Now().Add(-securityDashboardWindowHours * time.Hour)

	failedLoginCount, err := s.repo.CountByActions([]string{"auth.login.failed"}, since, schoolID)
	if err != nil {
		return nil, err
	}

	incidents, err := s.detectBruteForce(since, schoolID)
	if err != nil {
		return nil, err
	}

	resetRequested, err := s.repo.CountByActions([]string{"auth.password.reset.requested"}, since, schoolID)
	if err != nil {
		return nil, err
	}
	resetCompleted, err := s.repo.CountByActions([]string{"auth.password.reset.completed"}, since, schoolID)
	if err != nil {
		return nil, err
	}

	suspiciousLogs, err := s.repo.GetRecentByActions(suspiciousActivityActions, since, schoolID, suspiciousActivityLimit)
	if err != nil {
		return nil, err
	}

	return &dto.SecurityDashboardDTO{
		WindowHours:                 securityDashboardWindowHours,
		GeneratedAt:                 formatAPITime(time.Now()),
		FailedLoginCount:            failedLoginCount,
		BruteForceIncidents:         incidents,
		PasswordResetRequestedCount: resetRequested,
		PasswordResetCompletedCount: resetCompleted,
		SuspiciousActivities:        mapSuspiciousActivities(suspiciousLogs),
	}, nil
}

// detectBruteForce flags a target (an email OR a source IP) once its
// auth.login.failed attempts in the window include at least
// bruteForceMinFailures within any bruteForceWindow-length span — not
// merely bruteForceMinFailures spread across the whole lookback window,
// which would misclassify ordinary "forgot my password, a few tries over
// the day" activity as an incident.
//
// Both groupings run and are reported together, not one replacing the
// other — they represent different attack shapes: many failed attempts
// against one account (attacked from possibly-many sources) vs. many
// failed attempts from one source (possibly against many accounts).
// auth.login.failed never carries a user_id (identity is exactly what's
// unknown on a failed login, by design — see log.md), so email is the only
// account-side signal available; IP capture only started in Phase 11.5.1
// (AuthService.logLoginFailed), so log rows written before that never
// match the IP grouping — they're correctly treated as "IP unknown," not
// silently merged into whatever IP happens to be blank.
func (s *securityDashboardService) detectBruteForce(since time.Time, schoolID *string) ([]dto.BruteForceIncidentDTO, error) {
	emailGroups, err := s.repo.GroupFailedLoginsByEmail(since, bruteForceMinFailures, schoolID)
	if err != nil {
		return nil, err
	}

	incidents := make([]dto.BruteForceIncidentDTO, 0, len(emailGroups))
	for _, g := range emailGroups {
		times, err := s.repo.GetFailedLoginAttemptTimes(g.Email, since)
		if err != nil {
			return nil, err
		}
		if !hasDenseWindow(times, bruteForceMinFailures, bruteForceWindow) {
			continue
		}
		incidents = append(incidents, dto.BruteForceIncidentDTO{
			TargetType:    "email",
			Target:        g.Email,
			FailureCount:  g.FailureCount,
			LastAttemptAt: formatAPITime(g.LastAttemptAt),
		})
	}

	ipGroups, err := s.repo.GroupFailedLoginsByIP(since, bruteForceMinFailures, schoolID)
	if err != nil {
		return nil, err
	}
	for _, g := range ipGroups {
		times, err := s.repo.GetFailedLoginAttemptTimesByIP(g.IPAddress, since)
		if err != nil {
			return nil, err
		}
		if !hasDenseWindow(times, bruteForceMinFailures, bruteForceWindow) {
			continue
		}
		incidents = append(incidents, dto.BruteForceIncidentDTO{
			TargetType:    "ip",
			Target:        g.IPAddress,
			FailureCount:  g.FailureCount,
			LastAttemptAt: formatAPITime(g.LastAttemptAt),
		})
	}

	return incidents, nil
}

// hasDenseWindow reports whether times (ascending) contains any
// minCount-length run spanning at most window. times is already bounded
// to a small per-email slice (GetFailedLoginAttemptTimes caps at 200
// rows), so this plain sliding-window scan is cheap.
func hasDenseWindow(times []time.Time, minCount int, window time.Duration) bool {
	if len(times) < minCount {
		return false
	}
	for i := 0; i+minCount-1 < len(times); i++ {
		if times[i+minCount-1].Sub(times[i]) <= window {
			return true
		}
	}
	return false
}

func mapSuspiciousActivities(logs []*domain.Log) []dto.SuspiciousActivityDTO {
	result := make([]dto.SuspiciousActivityDTO, 0, len(logs))
	for _, l := range logs {
		item := dto.SuspiciousActivityDTO{
			LogID:     l.ID,
			Action:    l.Action,
			CreatedAt: formatAPITime(l.CreatedAt),
		}
		if l.Severity != nil {
			item.Severity = *l.Severity
		}
		if l.UserID != "" {
			item.UserID = l.UserID
			item.UserName = l.User.FullName
			item.UserEmail = l.User.Email
		}
		result = append(result, item)
	}
	return result
}
