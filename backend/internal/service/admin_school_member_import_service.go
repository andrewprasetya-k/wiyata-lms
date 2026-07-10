package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AdminSchoolMemberImportService interface {
	PreviewCSV(schoolID string, reader io.Reader) (*dto.AdminSchoolMemberImportPreviewResponseDTO, error)
	Commit(schoolID string, defaultPassword string, rows []dto.AdminSchoolMemberImportRowDTO) (*dto.AdminSchoolMemberImportCommitResponseDTO, error)
	ListMembers(schoolID string, search string, role string, includeDeleted bool, page int, limit int) (*dto.AdminSchoolMemberListResponseDTO, error)
	AddMember(schoolID string, input dto.AdminSchoolMemberCreateDTO) (*dto.AdminSchoolMemberResponseDTO, error)
	RemoveMember(schoolID string, schoolUserID string) error
	RestoreMember(schoolID string, schoolUserID string) error
}

type adminSchoolMemberImportService struct {
	db           *gorm.DB
	emailService EmailService
}

func NewAdminSchoolMemberImportService(db *gorm.DB, emailService EmailService) AdminSchoolMemberImportService {
	if emailService == nil {
		emailService = noopEmailService{}
	}
	return &adminSchoolMemberImportService{db: db, emailService: emailService}
}

type normalizedImportRow struct {
	RowNumber int
	FullName  string
	Email     string
	Role      string
	ClassCode string
	Errors    []string
}

type schoolMemberEmailJob struct {
	Email        string
	Role         string
	Notification string
}

func (s *adminSchoolMemberImportService) PreviewCSV(schoolID string, reader io.Reader) (*dto.AdminSchoolMemberImportPreviewResponseDTO, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	rows, err := s.parseCSV(content)
	if err != nil {
		return nil, err
	}

	return s.validateRows(schoolID, rows)
}

func (s *adminSchoolMemberImportService) Commit(schoolID string, defaultPassword string, rows []dto.AdminSchoolMemberImportRowDTO) (*dto.AdminSchoolMemberImportCommitResponseDTO, error) {
	if strings.TrimSpace(defaultPassword) == "" {
		return nil, errors.New("default password wajib diisi")
	}
	if len(rows) == 0 {
		return nil, errors.New("baris import wajib diisi")
	}

	normalizedRows := make([]normalizedImportRow, 0, len(rows))
	for _, row := range rows {
		normalizedRows = append(normalizedRows, normalizedImportRow{
			RowNumber: row.RowNumber,
			FullName:  strings.TrimSpace(row.FullName),
			Email:     strings.ToLower(strings.TrimSpace(row.Email)),
			Role:      strings.ToLower(strings.TrimSpace(row.Role)),
			ClassCode: strings.TrimSpace(row.ClassCode),
		})
	}

	preview, err := s.validateRows(schoolID, normalizedRows)
	if err != nil {
		return nil, err
	}
	if preview.InvalidCount > 0 {
		results := make([]dto.AdminSchoolMemberImportResultDTO, 0, len(preview.Rows))
		for _, row := range preview.Rows {
			reason := strings.Join(row.Errors, "; ")
			if reason == "" {
				reason = "Data belum valid."
			}
			results = append(results, dto.AdminSchoolMemberImportResultDTO{
				RowNumber: row.RowNumber,
				FullName:  row.FullName,
				Email:     row.Email,
				Role:      row.Role,
				ClassCode: row.ClassCode,
				Status:    "failed",
				Reason:    reason,
			})
		}
		return &dto.AdminSchoolMemberImportCommitResponseDTO{
			FailedCount: len(results),
			Results:     results,
		}, errors.New("data import masih memiliki baris yang tidak valid")
	}

	results := make([]dto.AdminSchoolMemberImportResultDTO, 0, len(preview.Rows))
	emailJobs := make([]schoolMemberEmailJob, 0, len(preview.Rows))
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, row := range preview.Rows {
			result := dto.AdminSchoolMemberImportResultDTO{
				RowNumber: row.RowNumber,
				FullName:  row.FullName,
				Email:     row.Email,
				Role:      row.Role,
				ClassCode: row.ClassCode,
				Status:    "imported",
			}

			user, createdUser, err := s.findOrCreateUser(tx, row.FullName, row.Email, defaultPassword)
			if err != nil {
				return fmt.Errorf("baris %d: %w", row.RowNumber, err)
			}
			schoolUser, membershipAction, err := s.findOrCreateSchoolUser(tx, schoolID, user.ID)
			if err != nil {
				return fmt.Errorf("baris %d: %w", row.RowNumber, err)
			}
			roleID, err := s.findRoleID(tx, row.Role)
			if err != nil {
				return fmt.Errorf("baris %d: %w", row.RowNumber, err)
			}
			roleAssigned, err := s.ensureRole(tx, schoolUser.ID, roleID)
			if err != nil {
				return fmt.Errorf("baris %d: %w", row.RowNumber, err)
			}
			classTouched := false
			if row.ClassCode != "" && row.Role == "student" {
				classID, err := s.findClassIDByCode(tx, schoolID, row.ClassCode)
				if err != nil {
					return fmt.Errorf("baris %d: %w", row.RowNumber, err)
				}
				classTouched, err = s.ensureActiveStudentEnrollment(tx, schoolID, schoolUser.ID, classID)
				if err != nil {
					return fmt.Errorf("baris %d: %w", row.RowNumber, err)
				}
			}

			result.UserCreated = createdUser
			result.MembershipAction = membershipAction
			result.EmailNotification = schoolMemberEmailNotification(createdUser, membershipAction, roleAssigned, classTouched)
			if membershipAction == "restored" {
				result.Reason = "Membership sekolah dipulihkan."
			}
			if !createdUser && membershipAction == "existing" && !roleAssigned && !classTouched {
				result.Status = "skipped"
				result.Reason = "Akun sudah menjadi warga sekolah dengan data yang sama."
				result.EmailNotification = ""
			}
			if result.Status == "imported" && result.EmailNotification != "" {
				emailJobs = append(emailJobs, schoolMemberEmailJob{
					Email:        row.Email,
					Role:         row.Role,
					Notification: result.EmailNotification,
				})
			}
			results = append(results, result)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.sendSchoolMemberEmailNotifications(schoolID, emailJobs)

	response := dto.AdminSchoolMemberImportCommitResponseDTO{Results: results}
	for _, result := range results {
		switch result.Status {
		case "imported":
			response.ImportedCount++
		case "skipped":
			response.SkippedCount++
		case "failed":
			response.FailedCount++
		}
	}
	return &response, nil
}

func (s *adminSchoolMemberImportService) ListMembers(schoolID string, search string, role string, includeDeleted bool, page int, limit int) (*dto.AdminSchoolMemberListResponseDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "" && !allowedSchoolMemberImportRoles[role] {
		return nil, errors.New("role filter tidak valid")
	}

	query := s.db.Model(&domain.SchoolUser{}).
		Preload("User").
		Preload("Roles.Role").
		Where("scu_sch_id = ?", schoolID)
	if includeDeleted {
		query = query.Unscoped()
	}
	if search = strings.TrimSpace(search); search != "" {
		searchTerm := "%" + search + "%"
		query = query.Joins("JOIN edv.users ON users.usr_id = school_users.scu_usr_id AND users.deleted_at IS NULL").
			Where("users.usr_nama_lengkap ILIKE ? OR users.usr_email ILIKE ?", searchTerm, searchTerm)
	}
	if role != "" {
		query = query.
			Joins("JOIN edv.user_roles ur ON ur.urol_scu_id = school_users.scu_id").
			Joins("JOIN edv.roles r ON r.rol_id = ur.urol_rol_id").
			Where("r.rol_name = ?", role)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var members []*domain.SchoolUser
	offset := (page - 1) * limit
	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&members).Error; err != nil {
		return nil, err
	}

	classCodes, err := s.classCodesBySchoolUser(schoolID, members)
	if err != nil {
		return nil, err
	}

	data := make([]dto.AdminSchoolMemberResponseDTO, 0, len(members))
	for _, member := range members {
		data = append(data, s.mapSchoolMember(member, classCodes[member.ID]))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &dto.AdminSchoolMemberListResponseDTO{
		Data:       data,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
	}, nil
}

func (s *adminSchoolMemberImportService) AddMember(schoolID string, input dto.AdminSchoolMemberCreateDTO) (*dto.AdminSchoolMemberResponseDTO, error) {
	row := normalizedImportRow{
		RowNumber: 1,
		FullName:  strings.TrimSpace(input.FullName),
		Email:     strings.ToLower(strings.TrimSpace(input.Email)),
		Role:      strings.ToLower(strings.TrimSpace(input.Role)),
		ClassCode: strings.TrimSpace(input.ClassCode),
	}
	preview, err := s.validateRows(schoolID, []normalizedImportRow{row})
	if err != nil {
		return nil, err
	}
	if preview.InvalidCount > 0 {
		return nil, errors.New(strings.Join(preview.Rows[0].Errors, "; "))
	}

	var response *dto.AdminSchoolMemberResponseDTO
	emailJobs := make([]schoolMemberEmailJob, 0, 1)
	err = s.db.Transaction(func(tx *gorm.DB) error {
		user, createdUser, err := s.findOrCreateUser(tx, row.FullName, row.Email, input.Password)
		if err != nil {
			return err
		}
		schoolUser, membershipAction, err := s.findOrCreateSchoolUser(tx, schoolID, user.ID)
		if err != nil {
			return err
		}
		roleID, err := s.findRoleID(tx, row.Role)
		if err != nil {
			return err
		}
		roleAssigned, err := s.ensureRole(tx, schoolUser.ID, roleID)
		if err != nil {
			return err
		}
		classTouched := false
		if row.ClassCode != "" && row.Role == "student" {
			classID, err := s.findClassIDByCode(tx, schoolID, row.ClassCode)
			if err != nil {
				return err
			}
			classTouched, err = s.ensureActiveStudentEnrollment(tx, schoolID, schoolUser.ID, classID)
			if err != nil {
				return err
			}
		}
		if err := tx.Preload("User").Preload("Roles.Role").Where("scu_id = ?", schoolUser.ID).First(schoolUser).Error; err != nil {
			return err
		}
		classCodes, err := s.classCodesBySchoolUserTx(tx, schoolID, []string{schoolUser.ID})
		if err != nil {
			return err
		}
		mapped := s.mapSchoolMember(schoolUser, classCodes[schoolUser.ID])
		mapped.UserCreated = createdUser
		mapped.MembershipAction = membershipAction
		mapped.EmailNotification = schoolMemberEmailNotification(createdUser, membershipAction, roleAssigned, classTouched)
		if mapped.EmailNotification != "" {
			emailJobs = append(emailJobs, schoolMemberEmailJob{
				Email:        row.Email,
				Role:         row.Role,
				Notification: mapped.EmailNotification,
			})
		}
		response = &mapped
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.sendSchoolMemberEmailNotifications(schoolID, emailJobs)
	return response, nil
}

func (s *adminSchoolMemberImportService) RemoveMember(schoolID string, schoolUserID string) error {
	leftAt := time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Unscoped().
			Model(&domain.SchoolUser{}).
			Where("scu_id = ? AND scu_sch_id = ? AND deleted_at IS NULL", schoolUserID, schoolID).
			Update("deleted_at", leftAt)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Model(&domain.Enrollment{}).
			Where("enr_scu_id = ? AND left_at IS NULL", schoolUserID).
			Update("left_at", leftAt).Error
	})
}

func (s *adminSchoolMemberImportService) RestoreMember(schoolID string, schoolUserID string) error {
	result := s.db.Unscoped().Model(&domain.SchoolUser{}).
		Where("scu_id = ? AND scu_sch_id = ?", schoolUserID, schoolID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func schoolMemberEmailNotification(createdUser bool, membershipAction string, roleAssigned bool, classTouched bool) string {
	if createdUser {
		return "account_created"
	}
	if membershipAction == "created" || membershipAction == "restored" || roleAssigned || classTouched {
		return "added_to_school"
	}
	return ""
}

func (s *adminSchoolMemberImportService) sendSchoolMemberEmailNotifications(schoolID string, jobs []schoolMemberEmailJob) {
	if len(jobs) == 0 {
		return
	}

	schoolName, err := s.findSchoolName(schoolID)
	if err != nil {
		fmt.Printf("[Email Warning] failed to load school for member notification school_id=%s error=%s\n", schoolID, err.Error())
		return
	}

	for _, job := range jobs {
		var sendErr error
		switch job.Notification {
		case "account_created":
			sendErr = s.emailService.SendSchoolMemberAccountCreated(job.Email, schoolName, job.Role)
		case "added_to_school":
			sendErr = s.emailService.SendSchoolMemberAddedToSchool(job.Email, schoolName, job.Role)
		default:
			continue
		}
		if sendErr != nil {
			fmt.Printf("[Email Warning] failed to send school member notification school_id=%s email=%s notification=%s error=%s\n", schoolID, maskEmail(job.Email), job.Notification, sendErr.Error())
		}
	}
}

func (s *adminSchoolMemberImportService) findSchoolName(schoolID string) (string, error) {
	var school domain.School
	if err := s.db.Select("sch_name").Where("sch_id = ?", schoolID).First(&school).Error; err != nil {
		return "", err
	}
	return school.Name, nil
}

