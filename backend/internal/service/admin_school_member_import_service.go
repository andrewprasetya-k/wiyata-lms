package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

var allowedSchoolMemberImportRoles = map[string]bool{
	"student": true,
	"teacher": true,
	"admin":   true,
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

func (s *adminSchoolMemberImportService) parseCSV(content []byte) ([]normalizedImportRow, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("file CSV tidak bisa dibaca: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("file CSV kosong")
	}

	header := make(map[string]int)
	for idx, value := range records[0] {
		header[strings.ToLower(strings.TrimSpace(value))] = idx
	}

	requiredHeaders := []string{"fullname", "email", "role"}
	for _, name := range requiredHeaders {
		if _, ok := header[name]; !ok {
			return nil, fmt.Errorf("kolom %s wajib ada", name)
		}
	}

	rows := make([]normalizedImportRow, 0, len(records)-1)
	classCodeIndex, hasClassCode := header["classcode"]
	if !hasClassCode {
		classCodeIndex = -1
	}

	for index, record := range records[1:] {
		rowNumber := index + 2
		if isEmptyCSVRecord(record) {
			continue
		}
		rows = append(rows, normalizedImportRow{
			RowNumber: rowNumber,
			FullName:  csvValue(record, header["fullname"]),
			Email:     strings.ToLower(csvValue(record, header["email"])),
			Role:      strings.ToLower(csvValue(record, header["role"])),
			ClassCode: csvValue(record, classCodeIndex),
		})
	}

	if len(rows) == 0 {
		return nil, errors.New("file CSV tidak memiliki baris data")
	}
	return rows, nil
}

func (s *adminSchoolMemberImportService) validateRows(schoolID string, rows []normalizedImportRow) (*dto.AdminSchoolMemberImportPreviewResponseDTO, error) {
	duplicateEmails := map[string]int{}
	classCodes := map[string]bool{}

	for _, row := range rows {
		if row.Email != "" {
			duplicateEmails[row.Email]++
		}
		if row.ClassCode != "" {
			classCodes[row.ClassCode] = true
		}
	}

	existingClassCodes, err := s.existingClassCodes(schoolID, classCodes)
	if err != nil {
		return nil, err
	}

	response := &dto.AdminSchoolMemberImportPreviewResponseDTO{
		Rows: make([]dto.AdminSchoolMemberImportRowDTO, 0, len(rows)),
	}

	for _, row := range rows {
		errorsForRow := append([]string{}, row.Errors...)
		if row.FullName == "" {
			errorsForRow = append(errorsForRow, "Nama lengkap wajib diisi.")
		}
		if row.Email == "" {
			errorsForRow = append(errorsForRow, "Email wajib diisi.")
		} else if _, err := mail.ParseAddress(row.Email); err != nil {
			errorsForRow = append(errorsForRow, "Format email tidak valid.")
		}
		if duplicateEmails[row.Email] > 1 {
			errorsForRow = append(errorsForRow, "Email duplikat di file import.")
		}
		if row.Role == "" {
			errorsForRow = append(errorsForRow, "Peran wajib diisi.")
		} else if !allowedSchoolMemberImportRoles[row.Role] {
			errorsForRow = append(errorsForRow, "Peran hanya boleh student, teacher, atau admin.")
		}
		if row.ClassCode != "" {
			if row.Role != "student" {
				errorsForRow = append(errorsForRow, "classCode hanya berlaku untuk peran student.")
			}
			if !existingClassCodes[row.ClassCode] {
				errorsForRow = append(errorsForRow, "Kode kelas tidak ditemukan di sekolah aktif.")
			}
		}

		status := "valid"
		if len(errorsForRow) > 0 {
			status = "invalid"
			response.InvalidCount++
		} else {
			response.ValidCount++
		}
		response.Rows = append(response.Rows, dto.AdminSchoolMemberImportRowDTO{
			RowNumber: row.RowNumber,
			FullName:  row.FullName,
			Email:     row.Email,
			Role:      row.Role,
			ClassCode: row.ClassCode,
			Status:    status,
			Errors:    errorsForRow,
		})
	}

	return response, nil
}

func (s *adminSchoolMemberImportService) existingClassCodes(schoolID string, classCodes map[string]bool) (map[string]bool, error) {
	result := map[string]bool{}
	if len(classCodes) == 0 {
		return result, nil
	}

	codes := make([]string, 0, len(classCodes))
	for code := range classCodes {
		codes = append(codes, code)
	}

	var found []string
	if err := s.db.Model(&domain.Class{}).
		Where("cls_sch_id = ? AND cls_code IN ? AND deleted_at IS NULL", schoolID, codes).
		Pluck("cls_code", &found).Error; err != nil {
		return nil, err
	}
	for _, code := range found {
		result[code] = true
	}
	return result, nil
}

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

func csvValue(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func isEmptyCSVRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
