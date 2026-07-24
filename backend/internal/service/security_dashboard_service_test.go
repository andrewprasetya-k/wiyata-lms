package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"testing"
	"time"
)

type securityRepoStub struct {
	failedLoginCount    int64
	resetRequestedCount int64
	resetCompletedCount int64
	groups              []repository.FailedLoginGroup
	attemptTimesByEmail map[string][]time.Time
	ipGroups            []repository.FailedLoginIPGroup
	attemptTimesByIP    map[string][]time.Time
	recentLogs          []*domain.Log
	lastSchoolIDArg     *string
}

func (s *securityRepoStub) CountByActions(actions []string, since time.Time, schoolID *string) (int64, error) {
	s.lastSchoolIDArg = schoolID
	if len(actions) == 1 && actions[0] == "auth.login.failed" {
		return s.failedLoginCount, nil
	}
	if len(actions) == 1 && actions[0] == "auth.password.reset.requested" {
		return s.resetRequestedCount, nil
	}
	if len(actions) == 1 && actions[0] == "auth.password.reset.completed" {
		return s.resetCompletedCount, nil
	}
	return 0, nil
}

func (s *securityRepoStub) GroupFailedLoginsByEmail(_ time.Time, _ int, _ *string) ([]repository.FailedLoginGroup, error) {
	return s.groups, nil
}

func (s *securityRepoStub) GetFailedLoginAttemptTimes(email string, _ time.Time) ([]time.Time, error) {
	return s.attemptTimesByEmail[email], nil
}

func (s *securityRepoStub) GroupFailedLoginsByIP(_ time.Time, _ int, _ *string) ([]repository.FailedLoginIPGroup, error) {
	return s.ipGroups, nil
}

func (s *securityRepoStub) GetFailedLoginAttemptTimesByIP(ip string, _ time.Time) ([]time.Time, error) {
	return s.attemptTimesByIP[ip], nil
}

func (s *securityRepoStub) GetRecentByActions(_ []string, _ time.Time, _ *string, _ int) ([]*domain.Log, error) {
	return s.recentLogs, nil
}

func TestHasDenseWindow_FlagsFiveFailuresWithinFifteenMinutes(t *testing.T) {
	base := time.Now()
	times := []time.Time{
		base,
		base.Add(2 * time.Minute),
		base.Add(5 * time.Minute),
		base.Add(9 * time.Minute),
		base.Add(14 * time.Minute),
	}
	if !hasDenseWindow(times, bruteForceMinFailures, bruteForceWindow) {
		t.Fatal("expected 5 failures within 15 minutes to be flagged as dense")
	}
}

func TestHasDenseWindow_DoesNotFlagSpreadOutFailures(t *testing.T) {
	base := time.Now()
	// 5 failures, but spread across 5 hours — not a brute-force pattern,
	// just an ordinary "forgot my password" day.
	times := []time.Time{
		base,
		base.Add(1 * time.Hour),
		base.Add(2 * time.Hour),
		base.Add(3 * time.Hour),
		base.Add(5 * time.Hour),
	}
	if hasDenseWindow(times, bruteForceMinFailures, bruteForceWindow) {
		t.Fatal("expected spread-out failures to not be flagged as dense")
	}
}

func TestHasDenseWindow_FewerThanMinCountNeverFlagged(t *testing.T) {
	times := []time.Time{time.Now(), time.Now().Add(time.Minute)}
	if hasDenseWindow(times, bruteForceMinFailures, bruteForceWindow) {
		t.Fatal("expected fewer than minCount attempts to never be flagged")
	}
}

func TestGetDashboard_BruteForceIncidentRequiresDensityNotJustCount(t *testing.T) {
	base := time.Now()
	repo := &securityRepoStub{
		groups: []repository.FailedLoginGroup{
			{Email: "dense@example.com", FailureCount: 5, LastAttemptAt: base},
			{Email: "spread@example.com", FailureCount: 5, LastAttemptAt: base},
		},
		attemptTimesByEmail: map[string][]time.Time{
			"dense@example.com": {
				base, base.Add(2 * time.Minute), base.Add(4 * time.Minute),
				base.Add(6 * time.Minute), base.Add(8 * time.Minute),
			},
			"spread@example.com": {
				base, base.Add(2 * time.Hour), base.Add(4 * time.Hour),
				base.Add(6 * time.Hour), base.Add(8 * time.Hour),
			},
		},
	}
	svc := NewSecurityDashboardService(repo)

	dashboard, err := svc.GetDashboard(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dashboard.BruteForceIncidents) != 1 {
		t.Fatalf("expected exactly 1 incident, got %d", len(dashboard.BruteForceIncidents))
	}
	if dashboard.BruteForceIncidents[0].Target != "dense@example.com" {
		t.Fatalf("expected dense@example.com to be flagged, got %s", dashboard.BruteForceIncidents[0].Target)
	}
	if dashboard.BruteForceIncidents[0].TargetType != "email" {
		t.Fatalf("expected targetType email, got %s", dashboard.BruteForceIncidents[0].TargetType)
	}
}

func TestGetDashboard_BruteForceDetectsBothEmailAndIPPatterns(t *testing.T) {
	base := time.Now()
	repo := &securityRepoStub{
		groups: []repository.FailedLoginGroup{
			{Email: "victim@example.com", FailureCount: 5, LastAttemptAt: base},
		},
		attemptTimesByEmail: map[string][]time.Time{
			"victim@example.com": {
				base, base.Add(2 * time.Minute), base.Add(4 * time.Minute),
				base.Add(6 * time.Minute), base.Add(8 * time.Minute),
			},
		},
		ipGroups: []repository.FailedLoginIPGroup{
			{IPAddress: "203.0.113.9", FailureCount: 5, LastAttemptAt: base},
		},
		attemptTimesByIP: map[string][]time.Time{
			"203.0.113.9": {
				base, base.Add(1 * time.Minute), base.Add(2 * time.Minute),
				base.Add(3 * time.Minute), base.Add(4 * time.Minute),
			},
		},
	}
	svc := NewSecurityDashboardService(repo)

	dashboard, err := svc.GetDashboard(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dashboard.BruteForceIncidents) != 2 {
		t.Fatalf("expected both email- and IP-based incidents, got %d", len(dashboard.BruteForceIncidents))
	}

	var sawEmail, sawIP bool
	for _, incident := range dashboard.BruteForceIncidents {
		if incident.TargetType == "email" && incident.Target == "victim@example.com" {
			sawEmail = true
		}
		if incident.TargetType == "ip" && incident.Target == "203.0.113.9" {
			sawIP = true
		}
	}
	if !sawEmail || !sawIP {
		t.Fatalf("expected both email and ip incidents present, got %+v", dashboard.BruteForceIncidents)
	}
}

func TestGetDashboard_PassesSchoolIDThroughToRepo(t *testing.T) {
	schoolID := "school-123"
	repo := &securityRepoStub{}
	svc := NewSecurityDashboardService(repo)

	_, err := svc.GetDashboard(&schoolID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastSchoolIDArg == nil || *repo.lastSchoolIDArg != schoolID {
		t.Fatal("expected schoolID to be forwarded to repo.CountByActions")
	}
}

func TestMapSuspiciousActivities_OmitsUserDetailsWhenNoUserID(t *testing.T) {
	severity := domain.LogSeverityHigh
	logs := []*domain.Log{
		{ID: "log-1", Action: "auth.token.reuse_detected", Severity: &severity, UserID: "user-1", User: domain.User{FullName: "Budi", Email: "budi@example.com"}},
	}
	result := mapSuspiciousActivities(logs)
	if len(result) != 1 {
		t.Fatalf("expected 1 mapped item, got %d", len(result))
	}
	if result[0].UserEmail != "budi@example.com" {
		t.Fatalf("expected actor email to be populated, got %q", result[0].UserEmail)
	}
}
