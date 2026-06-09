package repository

import (
	"backend/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AssignmentRepository interface {
	// Category
	CreateCategory(cat *domain.AssignmentCategory) error
	GetCategoriesBySchool(schoolID string) ([]*domain.AssignmentCategory, error)

	// Assignment
	CreateAssignment(asg *domain.Assignment) error
	GetAssignmentsBySubjectClass(subjectClassID string, search string, page int, limit int) ([]*domain.Assignment, int64, error)
	GetAssignmentByID(id string) (*domain.Assignment, error)
	GetAssignmentWithSubmissions(id string) (*domain.Assignment, error)
	GetAssignmentsWithSubmissionsBySubjectClass(subjectClassID string, schoolID string) ([]*domain.Assignment, error)
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

func (r *assignmentRepository) CountStudentsInClass(classID string) (int, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("enr_cls_id = ? AND enr_role = ?", classID, "student").
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
		Update("deleted_at", gorm.Expr("now()"))

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

		now := time.Now()
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
