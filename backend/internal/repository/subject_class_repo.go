package repository

import (
	"backend/internal/domain"
	"gorm.io/gorm"
)

type SubjectClassRepository interface {
	Create(scl *domain.SubjectClass) error
	GetByClass(classID string) ([]*domain.SubjectClass, error)
	GetTeachingByUserAndSchool(userID string, schoolID string) ([]TeacherSubjectClassRow, error)
	GetByID(id string) (*domain.SubjectClass, error)
	Update(scl *domain.SubjectClass) error
	Delete(id string) error
	CheckExists(classID, subjectID, schoolUserID string) (bool, error)
	CheckClassSubjectExists(classID string, subjectID string, excludeID string) (bool, error)
	HasSubjectClassContent(subjectClassID string, schoolID string) (bool, error)
	GetClassIDBySubjectClass(subjectClassID string) (string, error)
	TeacherTeachesInClass(schoolUserID string, classID string) (bool, error)
	UserTeachesClass(userID string, schoolID string, classID string) (bool, error)
	TeacherOwnsSubjectClass(userID string, schoolID string, subjectClassID string) (bool, error)
	TeacherOwnsClassSubject(userID string, schoolID string, classID string, subjectID string) (bool, error)
	ClassBelongsToSchool(classID string, schoolID string) (bool, error)
	SubjectBelongsToSchool(subjectID string, schoolID string) (bool, error)
	SubjectClassBelongsToSchool(subjectClassID string, schoolID string) (bool, error)
	SchoolUserBelongsToSchool(schoolUserID string, schoolID string) (bool, error)
	SchoolUserHasRole(schoolUserID string, roleName string) (bool, error)
	SchoolUserEnrolledInClassAsRole(schoolUserID string, classID string, schoolID string, role string) (bool, error)
	UserEnrolledInSubjectClassAsRole(userID string, schoolID string, subjectClassID string, role string) (bool, error)
}

type subjectClassRepository struct {
	db *gorm.DB
}

type TeacherSubjectClassRow struct {
	SubjectClassID     string `gorm:"column:subject_class_id"`
	ClassID            string `gorm:"column:class_id"`
	ClassName          string `gorm:"column:class_name"`
	ClassCode          string `gorm:"column:class_code"`
	SubjectID          string `gorm:"column:subject_id"`
	SubjectName        string `gorm:"column:subject_name"`
	SubjectCode        string `gorm:"column:subject_code"`
	SubjectColor       string `gorm:"column:subject_color"`
	StudentCount       int64  `gorm:"column:student_count"`
	MaterialCount      int64  `gorm:"column:material_count"`
	AssignmentCount    int64  `gorm:"column:assignment_count"`
	PendingSubmissions int64  `gorm:"column:pending_submissions"`
}

func NewSubjectClassRepository(db *gorm.DB) SubjectClassRepository {
	return &subjectClassRepository{db: db}
}

func (r *subjectClassRepository) Create(scl *domain.SubjectClass) error {
	return r.db.Create(scl).Error
}

func (r *subjectClassRepository) GetByClass(classID string) ([]*domain.SubjectClass, error) {
	var results []*domain.SubjectClass
	err := r.db.Preload("Subject").Preload("Teacher.User").
		Where("scl_cls_id = ?", classID).Find(&results).Error
	return results, err
}

func (r *subjectClassRepository) GetTeachingByUserAndSchool(userID string, schoolID string) ([]TeacherSubjectClassRow, error) {
	var results []TeacherSubjectClassRow
	err := r.db.Raw(`
		SELECT
			sc.scl_id AS subject_class_id,
			c.cls_id AS class_id,
			c.cls_title AS class_name,
			c.cls_code AS class_code,
				sub.sub_id AS subject_id,
				sub.sub_name AS subject_name,
				sub.sub_code AS subject_code,
				COALESCE(sub.sub_color, '') AS subject_color,
				COUNT(DISTINCT enr.enr_scu_id) AS student_count,
			COUNT(DISTINCT mat.mat_id) AS material_count,
			COUNT(DISTINCT asg.asg_id) AS assignment_count,
			COUNT(DISTINCT CASE WHEN asm.asm_id IS NULL THEN sbm.sbm_id END) AS pending_submissions
		FROM edv.subject_classes sc
		JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id
		JOIN edv.enrollments teacher_enr
			ON teacher_enr.enr_cls_id = c.cls_id
			AND teacher_enr.enr_scu_id = sc.scl_scu_id
			AND teacher_enr.enr_sch_id = ?
			AND teacher_enr.enr_role = 'teacher'
			AND teacher_enr.left_at IS NULL
		LEFT JOIN edv.enrollments enr
			ON enr.enr_cls_id = c.cls_id
			AND enr.enr_sch_id = ?
			AND enr.enr_role = 'student'
			AND enr.left_at IS NULL
		LEFT JOIN edv.materials mat
			ON mat.mat_scl_id = sc.scl_id
			AND mat.mat_sch_id = ?
			AND mat.deleted_at IS NULL
		LEFT JOIN edv.assignments asg
			ON asg.asg_scl_id = sc.scl_id
			AND asg.asg_sch_id = ?
			AND asg.deleted_at IS NULL
		LEFT JOIN edv.submissions sbm
			ON sbm.sbm_asg_id = asg.asg_id
			AND sbm.sbm_sch_id = ?
			AND sbm.deleted_at IS NULL
		LEFT JOIN edv.assessments asm
			ON asm.asm_sbm_id = sbm.sbm_id
		WHERE teacher_scu.scu_usr_id = ?
			AND teacher_scu.scu_sch_id = ?
			AND c.cls_sch_id = ?
			AND sub.sub_sch_id = ?
			AND c.deleted_at IS NULL
			GROUP BY sc.scl_id, c.cls_id, c.cls_title, c.cls_code, sub.sub_id, sub.sub_name, sub.sub_code, sub.sub_color
		ORDER BY c.cls_title ASC, sub.sub_name ASC
	`, schoolID, schoolID, schoolID, schoolID, schoolID, userID, schoolID, schoolID, schoolID).Scan(&results).Error
	return results, err
}

func (r *subjectClassRepository) GetByID(id string) (*domain.SubjectClass, error) {
	var scl domain.SubjectClass
	err := r.db.Preload("Subject").Preload("Teacher.User").Preload("Class").
		Where("scl_id = ?", id).First(&scl).Error
	return &scl, err
}

func (r *subjectClassRepository) Update(scl *domain.SubjectClass) error {
	result := r.db.Save(scl)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *subjectClassRepository) Delete(id string) error {
	result := r.db.Delete(&domain.SubjectClass{}, "scl_id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *subjectClassRepository) CheckExists(classID, subjectID, schoolUserID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SubjectClass{}).
		Where("scl_cls_id = ? AND scl_sub_id = ? AND scl_scu_id = ?", classID, subjectID, schoolUserID).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) CheckClassSubjectExists(classID string, subjectID string, excludeID string) (bool, error) {
	var count int64
	query := r.db.Model(&domain.SubjectClass{}).
		Where("scl_cls_id = ? AND scl_sub_id = ?", classID, subjectID)
	if excludeID != "" {
		query = query.Where("scl_id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) HasSubjectClassContent(subjectClassID string, schoolID string) (bool, error) {
	var materialCount int64
	if err := r.db.Table("edv.materials").
		Where("mat_scl_id = ? AND mat_sch_id = ? AND deleted_at IS NULL", subjectClassID, schoolID).
		Count(&materialCount).Error; err != nil {
		return false, err
	}
	if materialCount > 0 {
		return true, nil
	}

	var assignmentCount int64
	if err := r.db.Table("edv.assignments").
		Where("asg_scl_id = ? AND asg_sch_id = ? AND deleted_at IS NULL", subjectClassID, schoolID).
		Count(&assignmentCount).Error; err != nil {
		return false, err
	}
	return assignmentCount > 0, nil
}

func (r *subjectClassRepository) GetClassIDBySubjectClass(subjectClassID string) (string, error) {
	var classID string
	err := r.db.Model(&domain.SubjectClass{}).
		Where("scl_id = ?", subjectClassID).
		Pluck("scl_cls_id", &classID).Error
	return classID, err
}

func (r *subjectClassRepository) TeacherTeachesInClass(schoolUserID string, classID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subject_classes sc").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = sc.scl_cls_id AND e.enr_scu_id = sc.scl_scu_id").
		Where("sc.scl_scu_id = ? AND sc.scl_cls_id = ?", schoolUserID, classID).
		Where("e.enr_role = ? AND e.left_at IS NULL", "teacher").
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) UserTeachesClass(userID string, schoolID string, classID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subject_classes sc").
		Joins("JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = sc.scl_cls_id AND e.enr_scu_id = sc.scl_scu_id").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Where("teacher_scu.scu_usr_id = ? AND teacher_scu.scu_sch_id = ? AND teacher_scu.deleted_at IS NULL", userID, schoolID).
		Where("sc.scl_cls_id = ?", classID).
		Where("e.enr_sch_id = ? AND e.enr_role = ? AND e.left_at IS NULL", schoolID, "teacher").
		Where("c.cls_sch_id = ? AND c.deleted_at IS NULL", schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) ClassBelongsToSchool(classID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.classes").
		Where("cls_id = ? AND cls_sch_id = ? AND deleted_at IS NULL", classID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) SubjectBelongsToSchool(subjectID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subjects").
		Where("sub_id = ? AND sub_sch_id = ?", subjectID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) SubjectClassBelongsToSchool(subjectClassID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subject_classes sc").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Joins("JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL").
		Where("sc.scl_id = ?", subjectClassID).
		Where("c.cls_sch_id = ? AND sub.sub_sch_id = ? AND teacher_scu.scu_sch_id = ?", schoolID, schoolID, schoolID).
		Where("c.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) SchoolUserBelongsToSchool(schoolUserID string, schoolID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.school_users").
		Where("scu_id = ? AND scu_sch_id = ? AND deleted_at IS NULL", schoolUserID, schoolID).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) SchoolUserHasRole(schoolUserID string, roleName string) (bool, error) {
	var count int64
	err := r.db.Table("edv.user_roles ur").
		Joins("JOIN edv.roles r ON r.rol_id = ur.urol_rol_id").
		Where("ur.urol_scu_id = ? AND r.rol_name = ?", schoolUserID, roleName).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) SchoolUserEnrolledInClassAsRole(schoolUserID string, classID string, schoolID string, role string) (bool, error) {
	var count int64
	err := r.db.Table("edv.enrollments").
		Where("enr_scu_id = ? AND enr_cls_id = ? AND enr_sch_id = ? AND enr_role = ? AND left_at IS NULL", schoolUserID, classID, schoolID, role).
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) UserEnrolledInSubjectClassAsRole(userID string, schoolID string, subjectClassID string, role string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subject_classes sc").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = c.cls_id").
		Joins("JOIN edv.school_users scu ON scu.scu_id = e.enr_scu_id AND scu.deleted_at IS NULL").
		Where("sc.scl_id = ?", subjectClassID).
		Where("c.cls_sch_id = ? AND e.enr_sch_id = ? AND scu.scu_sch_id = ? AND scu.deleted_at IS NULL", schoolID, schoolID, schoolID).
		Where("scu.scu_usr_id = ? AND e.enr_role = ?", userID, role).
		Where("e.left_at IS NULL").
		Where("c.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) TeacherOwnsSubjectClass(userID string, schoolID string, subjectClassID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subject_classes sc").
		Joins("JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = sc.scl_cls_id AND e.enr_scu_id = sc.scl_scu_id").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Where("sc.scl_id = ?", subjectClassID).
		Where("teacher_scu.scu_usr_id = ? AND teacher_scu.scu_sch_id = ? AND teacher_scu.deleted_at IS NULL", userID, schoolID).
		Where("e.enr_sch_id = ? AND e.enr_role = ? AND e.left_at IS NULL", schoolID, "teacher").
		Where("c.cls_sch_id = ? AND sub.sub_sch_id = ?", schoolID, schoolID).
		Where("c.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r *subjectClassRepository) TeacherOwnsClassSubject(userID string, schoolID string, classID string, subjectID string) (bool, error) {
	var count int64
	err := r.db.Table("edv.subject_classes sc").
		Joins("JOIN edv.school_users teacher_scu ON teacher_scu.scu_id = sc.scl_scu_id AND teacher_scu.deleted_at IS NULL").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = sc.scl_cls_id AND e.enr_scu_id = sc.scl_scu_id").
		Joins("JOIN edv.classes c ON c.cls_id = sc.scl_cls_id").
		Joins("JOIN edv.subjects sub ON sub.sub_id = sc.scl_sub_id").
		Where("sc.scl_cls_id = ? AND sc.scl_sub_id = ?", classID, subjectID).
		Where("teacher_scu.scu_usr_id = ? AND teacher_scu.scu_sch_id = ? AND teacher_scu.deleted_at IS NULL", userID, schoolID).
		Where("e.enr_sch_id = ? AND e.enr_role = ? AND e.left_at IS NULL", schoolID, "teacher").
		Where("c.cls_sch_id = ? AND sub.sub_sch_id = ?", schoolID, schoolID).
		Where("c.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}
