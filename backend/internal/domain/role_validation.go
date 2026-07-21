package domain

import (
	"errors"
	"strings"
)

// ErrStudentTeacherCombination and ErrStudentAdminCombination are returned
// when a school membership's resulting role set would combine "student"
// with "teacher" or "admin".
var (
	ErrStudentTeacherCombination = errors.New("role combination of student and teacher is not allowed")
	ErrStudentAdminCombination   = errors.New("role combination of student and admin is not allowed")
)

// ValidateSchoolRoleCombination is the single source of truth for which
// school-level role combinations are permitted on one school_users
// membership. It lives in domain (rather than service or repository) so
// that every mechanism which can change a membership's role set can call
// it without an import cycle: role sync / single-role assign (rbacService),
// CSV import / direct member creation (adminSchoolMemberImportService),
// and invitation accept — both the public new-user path and the
// authenticated existing-user path (invitationRepository).
//
// Allowed: admin, teacher, student, admin+teacher.
// Forbidden: student+teacher, student+admin.
// super_admin is intentionally out of scope: it is a platform role never
// assigned through any of these school-level flows.
func ValidateSchoolRoleCombination(roleNames []string) error {
	has := make(map[string]bool, len(roleNames))
	for _, name := range roleNames {
		has[strings.ToLower(strings.TrimSpace(name))] = true
	}

	if has["student"] && has["teacher"] {
		return ErrStudentTeacherCombination
	}
	if has["student"] && has["admin"] {
		return ErrStudentAdminCombination
	}
	return nil
}
