package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// buildActorContext assembles a domain.ActorContext from data already placed
// into the gin context by existing auth/RBAC middleware (JWT user id, active
// school id, active school-membership id). No new middleware is introduced;
// scope must be supplied by the caller since it depends on which route is
// being handled, not on anything derivable from context alone.
func buildActorContext(c *gin.Context, scope string) domain.ActorContext {
	actor := domain.ActorContext{
		UserID: middleware.GetUserID(c),
		Scope:  scope,
	}
	if raw, exists := c.Get("school_id"); exists {
		if value, ok := raw.(string); ok && value != "" {
			actor.SchoolID = &value
		}
	}
	if raw, exists := c.Get("school_user_id"); exists {
		if value, ok := raw.(string); ok && value != "" {
			actor.SchoolUserID = &value
		}
	}
	return actor
}
