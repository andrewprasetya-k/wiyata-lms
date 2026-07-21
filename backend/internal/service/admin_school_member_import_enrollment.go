package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Transaction-scoped member/enrollment helpers for
// AdminSchoolMemberImportService. Split out of
// admin_school_member_import_service.go to keep the main file focused on
// orchestration; behavior is unchanged.

func (s *adminSchoolMemberImportService) findOrCreateUser(tx *gorm.DB, fullName string, email string, defaultPassword string) (*domain.User, bool, error) {
	var user domain.User
	err := tx.Where("LOWER(usr_email) = ?", strings.ToLower(email)).First(&user).Error
	if err == nil {
		if !user.IsActive {
			return nil, false, errors.New("akun global tidak aktif")
		}
		return &user, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, false, err
	}
	user = domain.User{
		FullName: strings.TrimSpace(fullName),
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Password: string(hashedPassword),
		IsActive: true,
	}
	if err := tx.Create(&user).Error; err != nil {
		return nil, false, err
	}
	return &user, true, nil
}

func (s *adminSchoolMemberImportService) findOrCreateSchoolUser(tx *gorm.DB, schoolID string, userID string) (*domain.SchoolUser, string, error) {
	var schoolUser domain.SchoolUser
	err := tx.Unscoped().Where("scu_sch_id = ? AND scu_usr_id = ?", schoolID, userID).First(&schoolUser).Error
	if err == nil {
		if schoolUser.DeletedAt.Valid {
			if err := tx.Unscoped().Model(&domain.SchoolUser{}).
				Where("scu_id = ?", schoolUser.ID).
				Update("deleted_at", nil).Error; err != nil {
				return nil, "", err
			}
			schoolUser.DeletedAt = gorm.DeletedAt{}
			return &schoolUser, "restored", nil
		}
		return &schoolUser, "existing", nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}
	schoolUser = domain.SchoolUser{
		SchoolID: schoolID,
		UserID:   userID,
	}
	if err := tx.Create(&schoolUser).Error; err != nil {
		return nil, "", err
	}
	return &schoolUser, "created", nil
}

func (s *adminSchoolMemberImportService) findRoleID(tx *gorm.DB, roleName string) (string, error) {
	var role domain.Role
	if err := tx.Where("rol_name = ?", roleName).First(&role).Error; err != nil {
		return "", err
	}
	return role.ID, nil
}

func (s *adminSchoolMemberImportService) validateRoleAddition(tx *gorm.DB, schoolUserID string, newRoleName string) error {
	var existingNames []string
	err := tx.Table("edv.user_roles").
		Select("edv.roles.rol_name").
		Joins("JOIN edv.roles ON edv.roles.rol_id = edv.user_roles.urol_rol_id").
		Where("edv.user_roles.urol_scu_id = ?", schoolUserID).
		Pluck("edv.roles.rol_name", &existingNames).Error
	if err != nil {
		return err
	}
	return domain.ValidateSchoolRoleCombination(append(existingNames, newRoleName))
}

func (s *adminSchoolMemberImportService) ensureRole(tx *gorm.DB, schoolUserID string, roleID string) (bool, error) {
	userRole := domain.UserRole{
		SchoolUserID: schoolUserID,
		RoleID:       roleID,
	}
	result := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "urol_scu_id"}, {Name: "urol_rol_id"}},
		DoNothing: true,
	}).Create(&userRole)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (s *adminSchoolMemberImportService) findClassIDByCode(tx *gorm.DB, schoolID string, classCode string) (string, error) {
	var class domain.Class
	if err := tx.Where("cls_sch_id = ? AND cls_code = ? AND deleted_at IS NULL", schoolID, classCode).First(&class).Error; err != nil {
		return "", err
	}
	return class.ID, nil
}

func (s *adminSchoolMemberImportService) ensureActiveStudentEnrollment(tx *gorm.DB, schoolID string, schoolUserID string, classID string) (bool, error) {
	var enrollment domain.Enrollment
	err := tx.Where("enr_scu_id = ? AND enr_cls_id = ?", schoolUserID, classID).First(&enrollment).Error
	if err == nil {
		if enrollment.LeftAt == nil && enrollment.Role == "student" {
			return false, nil
		}
		if err := tx.Model(&domain.Enrollment{}).
			Where("enr_id = ?", enrollment.ID).
			Updates(map[string]any{"enr_role": "student", "left_at": nil}).Error; err != nil {
			return false, err
		}
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	enrollment = domain.Enrollment{
		SchoolID:     schoolID,
		SchoolUserID: schoolUserID,
		ClassID:      classID,
		Role:         "student",
	}
	if err := tx.Create(&enrollment).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (s *adminSchoolMemberImportService) classCodesBySchoolUser(schoolID string, members []*domain.SchoolUser) (map[string][]string, error) {
	ids := make([]string, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.ID)
	}
	return s.classCodesBySchoolUserTx(s.db, schoolID, ids)
}

func (s *adminSchoolMemberImportService) classCodesBySchoolUserTx(tx *gorm.DB, schoolID string, schoolUserIDs []string) (map[string][]string, error) {
	result := map[string][]string{}
	if len(schoolUserIDs) == 0 {
		return result, nil
	}

	type classCodeRow struct {
		SchoolUserID string
		ClassCode    string
	}
	var rows []classCodeRow
	err := tx.Table("edv.enrollments e").
		Select("e.enr_scu_id AS school_user_id, c.cls_code AS class_code").
		Joins("JOIN edv.classes c ON c.cls_id = e.enr_cls_id").
		Where("e.enr_sch_id = ? AND e.enr_scu_id IN ? AND e.left_at IS NULL", schoolID, schoolUserIDs).
		Where("c.cls_sch_id = ? AND c.deleted_at IS NULL", schoolID).
		Order("c.cls_code ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.SchoolUserID] = append(result[row.SchoolUserID], row.ClassCode)
	}
	return result, nil
}

func (s *adminSchoolMemberImportService) mapSchoolMember(member *domain.SchoolUser, classCodes []string) dto.AdminSchoolMemberResponseDTO {
	roles := make([]string, 0, len(member.Roles))
	for _, role := range member.Roles {
		if role.Role.Name == "" || role.Role.Name == "super_admin" {
			continue
		}
		roles = append(roles, role.Role.Name)
	}
	var deletedAt *string
	if member.DeletedAt.Valid {
		formatted := formatAPITime(member.DeletedAt.Time)
		deletedAt = &formatted
	}
	return dto.AdminSchoolMemberResponseDTO{
		SchoolUserID: member.ID,
		UserID:       member.UserID,
		FullName:     member.User.FullName,
		Email:        member.User.Email,
		Roles:        roles,
		ClassCodes:   classCodes,
		CreatedAt:    formatAPITime(member.CreatedAt),
		DeletedAt:    deletedAt,
	}
}
