package repository

import (
	"time"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	// Student
	GetPendingAssignmentsCount(userID string) (int, error)
	GetUpcomingDeadlines(userID string, limit int) ([]map[string]interface{}, error)
	GetAverageScore(userID string) (float64, error)
	GetMaterialProgress(userID string) (completed int, total int, err error)

	// Teacher
	GetPendingReviewsCount(schoolUserID string) (int, error)
	GetTotalStudentsByTeacher(schoolUserID string) (int, error)
	GetSubmissionRateByTeacher(schoolUserID string) (float64, error)
	GetClassPerformance(schoolUserID string) ([]map[string]interface{}, error)

	// Admin
	GetSchoolStatistics(schoolID string) (map[string]int, error)
	GetEnrollmentTrends(schoolID string) ([]map[string]interface{}, error)
	GetRecentActivities(schoolID string, limit int) ([]map[string]interface{}, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

// Student Dashboard
func (r *dashboardRepository) GetPendingAssignmentsCount(userID string) (int, error) {
	var count int64
	err := r.db.Table("edv.assignments a").
		Joins("JOIN edv.subject_classes sc ON a.asg_scl_id = sc.scl_id").
		Joins("JOIN edv.enrollments e ON sc.scl_cls_id = e.enr_cls_id").
		Joins("JOIN edv.school_users su ON e.enr_scu_id = su.scu_id AND su.deleted_at IS NULL").
		Where("su.scu_usr_id = ? AND e.left_at IS NULL AND a.asg_deadline > ? AND a.deleted_at IS NULL", userID, time.Now()).
		Where("NOT EXISTS (SELECT 1 FROM edv.submissions s WHERE s.sbm_asg_id = a.asg_id AND s.sbm_usr_id = ? AND s.deleted_at IS NULL)", userID).
		Count(&count).Error
	return int(count), err
}

func (r *dashboardRepository) GetUpcomingDeadlines(userID string, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT 
			a.asg_id as assignment_id,
			a.asg_title as assignment_title,
			sub.sub_name as subject_name,
			a.asg_deadline as deadline,
			EXISTS(SELECT 1 FROM edv.submissions s WHERE s.sbm_asg_id = a.asg_id AND s.sbm_usr_id = ? AND s.deleted_at IS NULL) as is_submitted
		FROM edv.assignments a
		JOIN edv.subject_classes sc ON a.asg_scl_id = sc.scl_id
		JOIN edv.subjects sub ON sc.scl_sub_id = sub.sub_id
		JOIN edv.enrollments e ON sc.scl_cls_id = e.enr_cls_id
		JOIN edv.school_users su ON e.enr_scu_id = su.scu_id AND su.deleted_at IS NULL
		WHERE su.scu_usr_id = ? 
			AND e.left_at IS NULL
			AND a.asg_deadline > ?
			AND a.deleted_at IS NULL
		ORDER BY a.asg_deadline ASC
		LIMIT ?
	`, userID, userID, time.Now(), limit).Scan(&results).Error
	return results, err
}

func (r *dashboardRepository) GetAverageScore(userID string) (float64, error) {
	var avg float64
	err := r.db.Raw(`
		SELECT COALESCE(AVG(asm.asm_score), 0) as average
		FROM edv.assessments asm
		JOIN edv.submissions s ON asm.asm_sbm_id = s.sbm_id
		WHERE s.sbm_usr_id = ? AND s.deleted_at IS NULL
	`, userID).Scan(&avg).Error
	return avg, err
}

func (r *dashboardRepository) GetMaterialProgress(userID string) (completed int, total int, err error) {
	err = r.db.Raw(`
		SELECT 
			COUNT(CASE WHEN mp.map_status = 'completed' THEN 1 END) as completed,
			COUNT(*) as total
		FROM edv.materials m
		JOIN edv.subject_classes sc ON m.mat_scl_id = sc.scl_id
		JOIN edv.enrollments e ON sc.scl_cls_id = e.enr_cls_id
		JOIN edv.school_users su ON e.enr_scu_id = su.scu_id AND su.deleted_at IS NULL
		LEFT JOIN edv.material_progress mp ON m.mat_id = mp.map_mat_id AND mp.map_usr_id = ?
		WHERE su.scu_usr_id = ? AND e.left_at IS NULL AND m.deleted_at IS NULL
	`, userID, userID).Row().Scan(&completed, &total)
	return
}

// Teacher Dashboard
func (r *dashboardRepository) GetPendingReviewsCount(schoolUserID string) (int, error) {
	var count int64
	err := r.db.Table("edv.submissions s").
		Joins("JOIN edv.assignments a ON s.sbm_asg_id = a.asg_id").
		Joins("JOIN edv.subject_classes sc ON a.asg_scl_id = sc.scl_id").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = sc.scl_cls_id AND e.enr_scu_id = sc.scl_scu_id").
		Where("sc.scl_scu_id = ? AND e.enr_role = 'teacher' AND e.left_at IS NULL AND s.deleted_at IS NULL", schoolUserID).
		Where("NOT EXISTS (SELECT 1 FROM edv.assessments asm WHERE asm.asm_sbm_id = s.sbm_id)").
		Count(&count).Error
	return int(count), err
}

func (r *dashboardRepository) GetTotalStudentsByTeacher(schoolUserID string) (int, error) {
	var count int64
	err := r.db.Raw(`
		SELECT COUNT(DISTINCT e.enr_scu_id)
		FROM edv.enrollments e
		JOIN edv.subject_classes sc ON e.enr_cls_id = sc.scl_cls_id
		JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = sc.scl_cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id
		WHERE sc.scl_scu_id = ? AND e.enr_role = 'student' AND e.left_at IS NULL
			AND teacher_e.enr_role = 'teacher' AND teacher_e.left_at IS NULL
	`, schoolUserID).Scan(&count).Error
	return int(count), err
}

func (r *dashboardRepository) GetSubmissionRateByTeacher(schoolUserID string) (float64, error) {
	var rate float64
	err := r.db.Raw(`
		SELECT 
			COALESCE(
				(
					COUNT(DISTINCT CASE
						WHEN s.sbm_id IS NOT NULL THEN CONCAT(a.asg_id::text, ':', e.enr_scu_id::text)
					END)::float
					/
					NULLIF(COUNT(DISTINCT CONCAT(a.asg_id::text, ':', e.enr_scu_id::text)), 0)
				) * 100,
				0
			) as rate
		FROM edv.assignments a
		JOIN edv.subject_classes sc ON a.asg_scl_id = sc.scl_id
		JOIN edv.classes c ON c.cls_id = sc.scl_cls_id
		JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = sc.scl_cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id
		JOIN edv.enrollments e ON e.enr_cls_id = sc.scl_cls_id
			AND e.enr_role = 'student'
			AND e.left_at IS NULL
			AND e.enr_sch_id = teacher_e.enr_sch_id
		JOIN edv.school_users student_scu ON student_scu.scu_id = e.enr_scu_id
			AND student_scu.scu_sch_id = teacher_e.enr_sch_id
			AND student_scu.deleted_at IS NULL
		LEFT JOIN edv.submissions s ON a.asg_id = s.sbm_asg_id
			AND s.sbm_usr_id = student_scu.scu_usr_id
			AND s.sbm_sch_id = teacher_e.enr_sch_id
			AND s.deleted_at IS NULL
		WHERE sc.scl_scu_id = ?
			AND teacher_e.enr_role = 'teacher'
			AND teacher_e.left_at IS NULL
			AND c.cls_sch_id = teacher_e.enr_sch_id
			AND c.deleted_at IS NULL
			AND a.asg_sch_id = teacher_e.enr_sch_id
			AND a.deleted_at IS NULL
	`, schoolUserID).Scan(&rate).Error
	return rate, err
}

func (r *dashboardRepository) GetClassPerformance(schoolUserID string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT 
			c.cls_id as class_id,
			c.cls_title as class_name,
			sub.sub_name as subject_name,
			COALESCE(sub.sub_color, '') as subject_color,
			COALESCE(AVG(asm.asm_score), 0) as average_score,
			COUNT(DISTINCT e.enr_scu_id) as total_students,
			COALESCE(
				(
					COUNT(DISTINCT CASE
						WHEN s.sbm_id IS NOT NULL THEN CONCAT(a.asg_id::text, ':', e.enr_scu_id::text)
					END)::float
					/
					NULLIF(COUNT(DISTINCT CASE
						WHEN a.asg_id IS NOT NULL AND e.enr_scu_id IS NOT NULL THEN CONCAT(a.asg_id::text, ':', e.enr_scu_id::text)
					END), 0)
				) * 100,
				0
			) as submission_rate
		FROM edv.subject_classes sc
		JOIN edv.classes c ON sc.scl_cls_id = c.cls_id
		JOIN edv.subjects sub ON sc.scl_sub_id = sub.sub_id
		JOIN edv.enrollments teacher_e ON teacher_e.enr_cls_id = sc.scl_cls_id AND teacher_e.enr_scu_id = sc.scl_scu_id
		LEFT JOIN edv.enrollments e ON c.cls_id = e.enr_cls_id
			AND e.enr_role = 'student'
			AND e.left_at IS NULL
			AND e.enr_sch_id = teacher_e.enr_sch_id
		LEFT JOIN edv.school_users student_scu ON student_scu.scu_id = e.enr_scu_id
			AND student_scu.scu_sch_id = teacher_e.enr_sch_id
			AND student_scu.deleted_at IS NULL
		LEFT JOIN edv.assignments a ON sc.scl_id = a.asg_scl_id
			AND a.asg_sch_id = teacher_e.enr_sch_id
			AND a.deleted_at IS NULL
		LEFT JOIN edv.submissions s ON a.asg_id = s.sbm_asg_id
			AND s.sbm_usr_id = student_scu.scu_usr_id
			AND s.sbm_sch_id = teacher_e.enr_sch_id
			AND s.deleted_at IS NULL
		LEFT JOIN edv.assessments asm ON s.sbm_id = asm.asm_sbm_id
		WHERE sc.scl_scu_id = ?
			AND teacher_e.enr_role = 'teacher'
			AND teacher_e.left_at IS NULL
			AND c.cls_sch_id = teacher_e.enr_sch_id
			AND c.deleted_at IS NULL
		GROUP BY c.cls_id, c.cls_title, sub.sub_name, sub.sub_color
	`, schoolUserID).Scan(&results).Error
	return results, err
}

// Admin Dashboard
func (r *dashboardRepository) GetSchoolStatistics(schoolID string) (map[string]int, error) {
	stats := make(map[string]int)

	var totalStudents, totalTeachers, totalClasses, activeClasses int64

	r.db.Table("edv.school_users su").
		Joins("JOIN edv.enrollments e ON su.scu_id = e.enr_scu_id").
		Where("su.scu_sch_id = ? AND su.deleted_at IS NULL AND e.enr_role = 'student' AND e.left_at IS NULL", schoolID).
		Count(&totalStudents)

	r.db.Table("edv.school_users su").
		Joins("JOIN edv.subject_classes sc ON su.scu_id = sc.scl_scu_id").
		Joins("JOIN edv.enrollments e ON e.enr_cls_id = sc.scl_cls_id AND e.enr_scu_id = sc.scl_scu_id").
		Where("su.scu_sch_id = ? AND su.deleted_at IS NULL AND e.enr_role = 'teacher' AND e.left_at IS NULL", schoolID).
		Count(&totalTeachers)

	r.db.Table("edv.classes").
		Where("cls_sch_id = ? AND deleted_at IS NULL", schoolID).
		Count(&totalClasses)

	r.db.Table("edv.classes").
		Where("cls_sch_id = ? AND is_active = true AND deleted_at IS NULL", schoolID).
		Count(&activeClasses)

	stats["totalStudents"] = int(totalStudents)
	stats["totalTeachers"] = int(totalTeachers)
	stats["totalClasses"] = int(totalClasses)
	stats["activeClasses"] = int(activeClasses)

	return stats, nil
}

func (r *dashboardRepository) GetEnrollmentTrends(schoolID string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT 
			c.cls_title as class_name,
			COUNT(DISTINCT e.enr_scu_id) as total_enrolled,
			COUNT(DISTINCT CASE WHEN e.enr_role = 'teacher' THEN e.enr_scu_id END) as teachers,
			COUNT(DISTINCT CASE WHEN e.enr_role = 'student' THEN e.enr_scu_id END) as students
		FROM edv.classes c
		LEFT JOIN edv.enrollments e ON c.cls_id = e.enr_cls_id AND e.left_at IS NULL
		WHERE c.cls_sch_id = ? AND c.deleted_at IS NULL
		GROUP BY c.cls_id, c.cls_title
		ORDER BY total_enrolled DESC
	`, schoolID).Scan(&results).Error
	return results, err
}

func (r *dashboardRepository) GetRecentActivities(schoolID string, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT 
			u.usr_nama_lengkap as user_name,
			l.log_action as action,
			l.created_at as timestamp
		FROM edv.logs l
		JOIN edv.users u ON l.log_usr_id = u.usr_id
		WHERE l.log_sch_id = ?
		ORDER BY l.created_at DESC
		LIMIT ?
	`, schoolID, limit).Scan(&results).Error
	return results, err
}
