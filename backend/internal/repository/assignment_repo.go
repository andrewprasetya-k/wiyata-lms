package repository

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/utils"
	"errors"

	"gorm.io/gorm"
)

type AssignmentRepository interface {
	// Category
	CreateCategory(cat *domain.AssignmentCategory) error
	GetCategoriesBySchool(schoolID string) ([]*domain.AssignmentCategory, error)
	AssignmentCategoryBelongsToSchool(categoryID string, schoolID string) (bool, error)

	// Assignment
	CreateAssignment(asg *domain.Assignment) error
	GetAssignmentsBySubjectClass(subjectClassID string, search string, page int, limit int) ([]*domain.Assignment, int64, error)
	GetAssignmentByID(id string) (*domain.Assignment, error)
	GetAssignmentWithSubmissions(id string) (*domain.Assignment, error)
	GetAssignmentsWithSubmissionsBySubjectClass(subjectClassID string, schoolID string) ([]*domain.Assignment, error)
	GetTeacherSubmissionInbox(userID string, schoolID string) ([]dto.TeacherSubmissionInboxItemDTO, error)
	GetTeacherAssignmentInbox(userID string, schoolID string) ([]dto.TeacherAssignmentInboxItemDTO, error)
	GetStudentAssignmentInbox(userID string, schoolID string) ([]dto.StudentAssignmentInboxItemDTO, error)
	CountStudentsInClass(classID string) (int, error)
	GetClassIDBySubjectClass(subjectClassID string) (string, error)
	UpdateAssignment(asg *domain.Assignment) error
	DeleteAssignment(id string) error

	// Submission
	UpsertSubmission(sbm *domain.Submission) error
	GetSubmissionsByAssignment(asgID string) ([]*domain.Submission, error)
	GetSubmissionByID(id string) (*domain.Submission, error)
	GetMySubmissionByAssignment(assignmentID string, userID string, schoolID string) (*domain.Submission, error)
	UpdateSubmission(sbm *domain.Submission) error
	DeleteSubmission(id string) error

	// Assessment
	UpsertAssessment(asm *domain.Assessment) error
	GetAssessmentBySubmission(sbmID string) (*domain.Assessment, error)
	UpdateAssessment(asm *domain.Assessment) error
	DeleteAssessment(submissionID string) error

	// Weights
	SetWeight(weight *domain.AssessmentWeight) error
	GetWeightsBySubject(subID string) ([]*domain.AssessmentWeight, error)
	DeleteBySubject(subID string) error
	GetTotalWeightBySubject(subID string) (float64, error)
}

type assignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) CreateCategory(cat *domain.AssignmentCategory) error {
	return r.db.Create(cat).Error
}

func (r *assignmentRepository) GetCategoriesBySchool(schoolID string) ([]*domain.AssignmentCategory, error) {
	var cats []*domain.AssignmentCategory
	err := r.db.Where("asc_sch_id = ?", schoolID).Find(&cats).Error
	return cats, err
}

func (r *assignmentRepository) AssignmentCategoryBelongsToSchool(categoryID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.AssignmentCategory{}).
		Where("asc_id = ? AND asc_sch_id = ?", categoryID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *assignmentRepository) CreateAssignment(asg *domain.Assignment) error {
	return r.db.Create(asg).Error
}

func (r *assignmentRepository) GetAssignmentsBySubjectClass(subjectClassID string, search string, page int, limit int) ([]*domain.Assignment, int64, error) {
	var results []*domain.Assignment
	var total int64

	query := r.db.Model(&domain.Assignment{}).
		Preload("Category").
		Preload("SubjectClass.Subject").
		Where("asg_scl_id = ?", subjectClassID)

	// Search by title or description
	if search != "" {
		query = query.Where("asg_title ILIKE ? OR asg_desc ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&results).Error
	return results, total, err
}

func (r *assignmentRepository) GetAssignmentByID(id string) (*domain.Assignment, error) {
	var asg domain.Assignment
	err := r.db.Preload("Category").Preload("SubjectClass.Subject").
		Where("asg_id = ?", id).First(&asg).Error
	return &asg, err
}

func (r *assignmentRepository) GetAssignmentWithSubmissions(id string) (*domain.Assignment, error) {
	var asg domain.Assignment
	err := r.db.Preload("Category").
		Preload("SubjectClass.Subject").
		Preload("SubjectClass.Class").
		Preload("Submissions.User").
		Preload("Submissions.Assessment.Assessor").
		Where("asg_id = ?", id).First(&asg).Error
	return &asg, err
}

func (r *assignmentRepository) GetAssignmentsWithSubmissionsBySubjectClass(subjectClassID string, schoolID string) ([]*domain.Assignment, error) {
	var assignments []*domain.Assignment
	err := r.db.Preload("Category").
		Preload("SubjectClass.Subject").
		Preload("SubjectClass.Teacher.User").
		Preload("Submissions", func(db *gorm.DB) *gorm.DB {
			return db.Where("sbm_sch_id = ?", schoolID).Order("submitted_at asc")
		}).
		Preload("Submissions.User").
		Preload("Submissions.Assessment.Assessor").
		Where("asg_scl_id = ? AND asg_sch_id = ?", subjectClassID, schoolID).
		Order("created_at desc").
		Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) GetTeacherSubmissionInbox(userID string, schoolID string) ([]dto.TeacherSubmissionInboxItemDTO, error) {
	var rows []dto.TeacherSubmissionInboxItemDTO
	err := r.db.Table("edv.assignments a").
		Select(`
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id,
			a.asg_title AS assignment_title,
			sub.sub_name AS subject_name,
			sub.sub_code AS subject_code,
			COALESCE(sub.sub_color, '') AS subject_color,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
			a.asg_deadline AS deadline,
			COUNT(s.sbm_id) AS submission_count,
			COUNT(CASE WHEN asm.asm_sbm_id IS NULL THEN s.sbm_id END) AS pending_count,
			COUNT(CASE WHEN asm.asm_sbm_id IS NOT NULL THEN s.sbm_id END) AS graded_count,
			COUNT(CASE WHEN a.asg_deadline IS NOT NULL AND s.submitted_at > a.asg_deadline THEN s.sbm_id END) AS late_count
		`).
		Joins("JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Joins("JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL").
		Joins("JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = sc.scl_cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id").
		Joins("LEFT JOIN edv.submissions s ON s.sbm_asg_id = a.asg_id AND s.sbm_sch_id = ? AND s.deleted_at IS NULL", schoolID).
		Joins("LEFT JOIN (SELECT DISTINCT asm_sbm_id FROM edv.assessments) asm ON asm.asm_sbm_id = s.sbm_id").
		Where("a.asg_sch_id = ? AND a.deleted_at IS NULL", schoolID).
		Where("teacher_scu.scu_usr_id = ? AND teacher_scu.scu_sch_id = ? AND teacher_scu.deleted_at IS NULL", userID, schoolID).
		Where("teacher_e.enr_sch_id = ? AND teacher_e.enr_role = ? AND teacher_e.left_at IS NULL", schoolID, "teacher").
		Where("c.cls_sch_id = ? AND c.deleted_at IS NULL", schoolID).
		Where("sub.sub_sch_id = ?", schoolID).
		Group("a.asg_id, sc.scl_id, a.asg_title, sub.sub_name, sub.sub_code, sub.sub_color, c.cls_title, c.cls_code, a.asg_deadline").
		Having("COUNT(s.sbm_id) > 0").
		Order("pending_count DESC, a.asg_deadline ASC NULLS LAST, a.asg_title ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *assignmentRepository) GetTeacherAssignmentInbox(userID string, schoolID string) ([]dto.TeacherAssignmentInboxItemDTO, error) {
	var rows []dto.TeacherAssignmentInboxItemDTO
	err := r.db.Table("edv.assignments a").
		Select(`
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id,
				a.asg_title AS assignment_title,
				sub.sub_name AS subject_name,
				sub.sub_code AS subject_code,
				COALESCE(sub.sub_color, '') AS subject_color,
				c.cls_title AS class_name,
			c.cls_code AS class_code,
			COALESCE(ac.asc_name, '') AS category_name,
			a.asg_deadline AS deadline,
			COUNT(s.sbm_id) AS submission_count,
			COUNT(CASE WHEN asm.asm_sbm_id IS NULL THEN s.sbm_id END) AS pending_count,
			COUNT(CASE WHEN asm.asm_sbm_id IS NOT NULL THEN s.sbm_id END) AS graded_count,
			COUNT(CASE WHEN a.asg_deadline IS NOT NULL AND s.submitted_at > a.asg_deadline THEN s.sbm_id END) AS late_count
		`).
		Joins("JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Joins("JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL").
		Joins("JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = sc.scl_cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id").
		Joins("LEFT JOIN edv.assignment_categories ac ON ac.asc_id = a.asg_asc_id").
		Joins("LEFT JOIN edv.submissions s ON s.sbm_asg_id = a.asg_id AND s.sbm_sch_id = ? AND s.deleted_at IS NULL", schoolID).
		Joins("LEFT JOIN (SELECT DISTINCT asm_sbm_id FROM edv.assessments) asm ON asm.asm_sbm_id = s.sbm_id").
		Where("a.asg_sch_id = ? AND a.deleted_at IS NULL", schoolID).
		Where("teacher_scu.scu_usr_id = ? AND teacher_scu.scu_sch_id = ? AND teacher_scu.deleted_at IS NULL", userID, schoolID).
		Where("teacher_e.enr_sch_id = ? AND teacher_e.enr_role = ? AND teacher_e.left_at IS NULL", schoolID, "teacher").
		Where("c.cls_sch_id = ? AND c.deleted_at IS NULL", schoolID).
		Where("sub.sub_sch_id = ?", schoolID).
		Group("a.asg_id, sc.scl_id, a.asg_title, sub.sub_name, sub.sub_code, sub.sub_color, c.cls_title, c.cls_code, ac.asc_name, a.asg_deadline").
		Order("pending_count DESC, a.asg_deadline ASC NULLS LAST, a.asg_title ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *assignmentRepository) GetStudentAssignmentInbox(userID string, schoolID string) ([]dto.StudentAssignmentInboxItemDTO, error) {
	var rows []dto.StudentAssignmentInboxItemDTO
	now := utils.NowJakarta()
	err := r.db.Table("edv.assignments a").
		Select(`
			a.asg_id AS assignment_id,
			sc.scl_id AS subject_class_id,
				a.asg_title AS assignment_title,
				sub.sub_name AS subject_name,
				sub.sub_code AS subject_code,
				COALESCE(sub.sub_color, '') AS subject_color,
				c.cls_title AS class_name,
			c.cls_code AS class_code,
			COALESCE(ac.asc_name, '') AS category_name,
			a.asg_deadline AS deadline,
			s.sbm_id AS submission_id,
			s.submitted_at AS submitted_at,
			asm.asm_score AS score,
			(s.sbm_id IS NOT NULL) AS is_submitted,
			(asm.asm_sbm_id IS NOT NULL) AS is_graded,
			(a.asg_deadline IS NOT NULL AND a.asg_deadline < ? AND s.sbm_id IS NULL) AS is_overdue,
			(a.asg_deadline IS NOT NULL AND s.submitted_at IS NOT NULL AND s.submitted_at > a.asg_deadline) AS is_submitted_late
		`, now).
		Joins("JOIN edv.subject_classes sc ON sc.scl_id = a.asg_scl_id").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id").
		Joins("JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id AND scu.deleted_at IS NULL").
		Joins("LEFT JOIN edv.assignment_categories ac ON ac.asc_id = a.asg_asc_id").
		Joins(`LEFT JOIN LATERAL (
			SELECT *
			FROM edv.submissions latest_s
			WHERE latest_s.sbm_asg_id = a.asg_id
				AND latest_s.sbm_usr_id = ?
				AND latest_s.sbm_sch_id = ?
				AND latest_s.deleted_at IS NULL
			ORDER BY latest_s.submitted_at DESC, latest_s.sbm_id DESC
			LIMIT 1
		) s ON true`, userID, schoolID).
		Joins(`LEFT JOIN LATERAL (
			SELECT *
			FROM edv.assessments latest_asm
			WHERE latest_asm.asm_sbm_id = s.sbm_id
			ORDER BY latest_asm.assessed_at DESC, latest_asm.asm_id DESC
			LIMIT 1
		) asm ON true`).
		Where("a.asg_sch_id = ? AND a.deleted_at IS NULL", schoolID).
		Where("c.cls_sch_id = ? AND c.deleted_at IS NULL", schoolID).
		Where("sub.sub_sch_id = ?", schoolID).
		Where("e.enr_sch_id = ? AND e.enr_role = ? AND e.left_at IS NULL", schoolID, "student").
		Where("scu.scu_usr_id = ? AND scu.scu_sch_id = ? AND scu.deleted_at IS NULL", userID, schoolID).
		Order("is_overdue DESC, is_submitted ASC, a.asg_deadline ASC NULLS LAST, a.asg_title ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *assignmentRepository) CountStudentsInClass(classID string) (int, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("enr_cls_id = ? AND enr_role = ? AND left_at IS NULL", classID, "student").
		Count(&count).Error
	return int(count), err
}

func (r *assignmentRepository) GetClassIDBySubjectClass(subjectClassID string) (string, error) {
	var classID string
	err := r.db.Model(&domain.SubjectClass{}).
		Where("scl_id = ?", subjectClassID).
		Pluck("scl_cls_id", &classID).Error
	return classID, err
}

func (r *assignmentRepository) UpdateAssignment(asg *domain.Assignment) error {
	result := r.db.Model(&domain.Assignment{}).Where("asg_id = ?", asg.ID).Updates(asg)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *assignmentRepository) DeleteAssignment(id string) error {
	result := r.db.Where("asg_id = ?", id).Delete(&domain.Assignment{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *assignmentRepository) UpsertSubmission(sbm *domain.Submission) error {
	var existing domain.Submission
	// cek apakah user sudah pernah submit assignment ini sebelumnya
	err := r.db.Unscoped().Where("sbm_asg_id = ? AND sbm_usr_id = ?", sbm.AssignmentID, sbm.UserID).First(&existing).Error

	if err == nil {
		sbm.ID = existing.ID
		sbm.DeletedAt = gorm.DeletedAt{} //reset deleted_at
		return r.db.Save(sbm).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(sbm).Error
	}

	return err
}

func (r *assignmentRepository) GetSubmissionsByAssignment(asgID string) ([]*domain.Submission, error) {
	var results []*domain.Submission
	err := r.db.Preload("User").Where("sbm_asg_id = ?", asgID).Order("submitted_at asc").Find(&results).Error
	return results, err
}

func (r *assignmentRepository) GetSubmissionByID(id string) (*domain.Submission, error) {
	var sbm domain.Submission
	err := r.db.Preload("User").Preload("Assessment.Assessor").Where("sbm_id = ?", id).First(&sbm).Error
	return &sbm, err
}

func (r *assignmentRepository) GetMySubmissionByAssignment(assignmentID string, userID string, schoolID string) (*domain.Submission, error) {
	var sbm domain.Submission
	query := r.db.Preload("User").
		Preload("Assessment.Assessor").
		Where("sbm_asg_id = ? AND sbm_usr_id = ?", assignmentID, userID)

	if schoolID != "" {
		query = query.Where("sbm_sch_id = ?", schoolID)
	}

	err := query.First(&sbm).Error
	return &sbm, err
}

func (r *assignmentRepository) UpdateSubmission(sbm *domain.Submission) error {
	result := r.db.Model(&domain.Submission{}).Where("sbm_id = ?", sbm.ID).Updates(sbm)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *assignmentRepository) DeleteSubmission(id string) error {
	// Gunakan gorm.Expr agar "now()" dianggap sebagai fungsi SQL, bukan string biasa
	result := r.db.Model(&domain.Submission{}).
		Where("sbm_id = ?", id).
		Update("deleted_at", utils.NowJakarta())

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *assignmentRepository) UpsertAssessment(asm *domain.Assessment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing []domain.Assessment
		if err := tx.
			Where("asm_sbm_id = ?", asm.SubmissionID).
			Order("assessed_at desc, asm_id desc").
			Find(&existing).Error; err != nil {
			return err
		}

		now := utils.NowJakarta()
		if len(existing) == 0 {
			asm.AssessedAt = now
			return tx.Create(asm).Error
		}

		keepID := existing[0].ID
		if err := tx.Model(&domain.Assessment{}).
			Where("asm_id = ?", keepID).
			Updates(map[string]any{
				"asm_score":    asm.Score,
				"asm_feedback": asm.Feedback,
				"assessed_by":  asm.AssessedBy,
				"assessed_at":  now,
				"asm_sbm_id":   asm.SubmissionID,
			}).Error; err != nil {
			return err
		}

		if len(existing) > 1 {
			duplicateIDs := make([]string, 0, len(existing)-1)
			for _, item := range existing[1:] {
				duplicateIDs = append(duplicateIDs, item.ID)
			}
			if err := tx.Where("asm_id IN ?", duplicateIDs).Delete(&domain.Assessment{}).Error; err != nil {
				return err
			}
		}

		asm.ID = keepID
		asm.AssessedAt = now
		return nil
	})
}

func (r *assignmentRepository) GetAssessmentBySubmission(sbmID string) (*domain.Assessment, error) {
	var asm domain.Assessment
	err := r.db.Preload("Assessor").Where("asm_sbm_id = ?", sbmID).First(&asm).Error
	return &asm, err
}

func (r *assignmentRepository) UpdateAssessment(asm *domain.Assessment) error {
	result := r.db.Model(&domain.Assessment{}).Where("asm_sbm_id = ?", asm.SubmissionID).Updates(asm)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *assignmentRepository) DeleteAssessment(submissionID string) error {
	result := r.db.Where("asm_sbm_id = ?", submissionID).Delete(&domain.Assessment{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *assignmentRepository) SetWeight(weight *domain.AssessmentWeight) error {
	return r.db.Save(weight).Error
}

func (r *assignmentRepository) GetWeightsBySubject(subID string) ([]*domain.AssessmentWeight, error) {
	var weights []*domain.AssessmentWeight
	err := r.db.Preload("Category").Where("asw_sub_id = ?", subID).Find(&weights).Error
	return weights, err
}

func (r *assignmentRepository) DeleteBySubject(subID string) error {
	result := r.db.Where("asw_sub_id = ?", subID).Delete(&domain.AssessmentWeight{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *assignmentRepository) GetTotalWeightBySubject(subID string) (float64, error) {
	var total float64
	err := r.db.Model(&domain.AssessmentWeight{}).Where("asw_sub_id = ?", subID).Select("SUM(asw_weight)").Scan(&total).Error
	return total, err
}
