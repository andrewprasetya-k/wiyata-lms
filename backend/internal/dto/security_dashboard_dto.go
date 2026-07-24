package dto

// SecurityDashboardDTO is the response for both
// GET /dashboard/admin/:schoolId/security and
// GET /dashboard/super-admin/security — the school pin (or its absence,
// for super_admin) is applied entirely on the backend before this shape
// is built, so the two handlers share one DTO and one service method.
type SecurityDashboardDTO struct {
	WindowHours                 int                     `json:"windowHours"`
	GeneratedAt                 string                  `json:"generatedAt"`
	FailedLoginCount            int64                   `json:"failedLoginCount"`
	BruteForceIncidents         []BruteForceIncidentDTO `json:"bruteForceIncidents"`
	PasswordResetRequestedCount int64                   `json:"passwordResetRequestedCount"`
	PasswordResetCompletedCount int64                   `json:"passwordResetCompletedCount"`
	SuspiciousActivities        []SuspiciousActivityDTO `json:"suspiciousActivities"`
}

// BruteForceIncidentDTO is one target (an email address OR a source IP)
// whose failed-login attempts crossed the brute-force density threshold
// (see SecurityDashboardService — currently 5 failures within any
// 15-minute span). Phase 11.5 only grouped by target email; Phase 11.5.1
// added IP-based grouping once auth.login.failed started capturing the
// caller's IP — the two represent different attack shapes (one account
// attacked from many sources vs. many accounts attacked from one source),
// so both are surfaced rather than one replacing the other. TargetType
// disambiguates which grouping produced this row.
type BruteForceIncidentDTO struct {
	TargetType    string `json:"targetType"` // "email" or "ip"
	Target        string `json:"target"`
	FailureCount  int64  `json:"failureCount"`
	LastAttemptAt string `json:"lastAttemptAt"`
}

// SuspiciousActivityDTO is one row of the combined
// reuse-detected / recovery-code-used / repeated-MFA-failure feed. UserName
// and UserEmail are populated from the actor's user_id when the underlying
// action has one (all three do) — this is the same actor detail level
// already exposed to admin/super_admin on the existing audit log viewer
// (AuditLogListItem.actorEmail), not a new disclosure.
type SuspiciousActivityDTO struct {
	LogID     string `json:"logId"`
	Action    string `json:"action"`
	Severity  string `json:"severity"`
	UserID    string `json:"userId,omitempty"`
	UserName  string `json:"userName,omitempty"`
	UserEmail string `json:"userEmail,omitempty"`
	CreatedAt string `json:"createdAt"`
}
