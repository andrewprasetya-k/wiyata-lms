package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"net/mail"
)

// Row-validation helpers for AdminSchoolMemberImportService. Split out of
// admin_school_member_import_service.go to keep the main file focused on
// orchestration; behavior is unchanged.

var allowedSchoolMemberImportRoles = map[string]bool{
	"student": true,
	"teacher": true,
	"admin":   true,
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
