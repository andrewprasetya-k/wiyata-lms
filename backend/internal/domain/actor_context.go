package domain

type ActorContext struct {
	UserID       string
	SchoolUserID *string
	SchoolID     *string
	Scope        string
	// IPAddress is the caller's request IP, when the call site has one
	// available (see RefreshTokenMetadata) — nil for actions with no
	// request context (e.g. the retention cleanup job) or where the
	// call site hasn't been wired up to pass it through yet.
	IPAddress *string
}
