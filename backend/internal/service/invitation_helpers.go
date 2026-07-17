package service

import (
	"os"
	"strings"
)

// buildInvitationAcceptURL and maskEmail are shared by the invitation-style
// email flows (school member invitation, school member import) — moved out
// of the now-deleted school_registration_request_service.go, which is where
// they used to live, into their own file so they don't get lost again.

func buildInvitationAcceptURL(rawToken string) string {
	path := "/invite/" + rawToken
	publicURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if publicURL == "" {
		return path
	}
	return publicURL + path
}

func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" {
		return "unknown"
	}
	local := parts[0]
	if len(local) <= 2 {
		return local[:1] + "***@" + parts[1]
	}
	return local[:2] + "***@" + parts[1]
}
