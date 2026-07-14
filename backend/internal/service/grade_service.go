package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

var ErrStudentNotEnrolledInClass = errors.New("student is not enrolled in this class")

type GradeService interface {
	ConfigureWeights(req *dto.ConfigureWeightsDTO, schoolID string) error
	GetWeightsBySubject(subjectID string, schoolID string) (*dto.WeightResponseDTO, error)
	CalculateFinalGrade(studentID string, subjectID string) (*dto.GradeReportDTO, error)
	GetClassGradeReport(classID, subjectID, schoolID string) (*dto.ClassGradeReportDTO, error)
	GetMyGradebookByClass(userID string, schoolID string, classID string) (*dto.MyGradebookResponseDTO, error)
	GetStudentGradeDetail(classID, subjectID, studentID, schoolID string) (*dto.StudentGradeDetailDTO, error)
	GetStudentReport(classID, studentID, schoolID string) (*dto.StudentReportDTO, error)
}

type gradeService struct {
	weightRepo  repository.AssessmentWeightRepository
	gradeRepo   repository.GradeRepository
	subjectRepo repository.SubjectRepository
	classRepo   repository.ClassRepository
	userRepo    repository.UserRepository
}

func NewGradeService(
	weightRepo repository.AssessmentWeightRepository,
	gradeRepo repository.GradeRepository,
	subjectRepo repository.SubjectRepository,
	classRepo repository.ClassRepository,
	userRepo repository.UserRepository,
) GradeService {
	return &gradeService{
		weightRepo:  weightRepo,
		gradeRepo:   gradeRepo,
		subjectRepo: subjectRepo,
		classRepo:   classRepo,
		userRepo:    userRepo,
	}
}

func (s *gradeService) ConfigureWeights(req *dto.ConfigureWeightsDTO, schoolID string) error {
	if len(req.Weights) == 0 {
		return fmt.Errorf("assessment weights are required")
	}

	if err := s.ensureWeightSubjectInSchool(req.SubjectID, schoolID); err != nil {
		return err
	}

	totalWeight := 0.0
	categoryIDs := make([]string, 0, len(req.Weights))
	seenCategories := make(map[string]struct{}, len(req.Weights))
	for _, w := range req.Weights {
		if w.Weight == nil {
			return fmt.Errorf("assessment weight is required")
		}
		weightValue := *w.Weight
		if weightValue < 0 || weightValue > 100 {
			return fmt.Errorf("assessment weight must be between 0 and 100")
		}
		if _, exists := seenCategories[w.CategoryID]; exists {
			return fmt.Errorf("duplicate assessment category in weights")
		}
		seenCategories[w.CategoryID] = struct{}{}
		categoryIDs = append(categoryIDs, w.CategoryID)
		totalWeight += weightValue
	}

	if math.Abs(totalWeight-100.0) > 0.01 {
		return fmt.Errorf("total weight must be 100, got %.2f", totalWeight)
	}

	if err := s.ensureWeightCategoriesInSchool(categoryIDs, schoolID); err != nil {
		return err
	}

	weights := make([]*domain.AssessmentWeight, 0, len(req.Weights))
	for _, w := range req.Weights {
		weights = append(weights, &domain.AssessmentWeight{
			SubjectID:  req.SubjectID,
			CategoryID: w.CategoryID,
			Weight:     *w.Weight,
		})
	}

	return s.weightRepo.ReplaceBySubject(req.SubjectID, weights)
}

func (s *gradeService) GetWeightsBySubject(subjectID string, schoolID string) (*dto.WeightResponseDTO, error) {
	if err := s.ensureWeightSubjectInSchool(subjectID, schoolID); err != nil {
		return nil, err
	}

	weights, err := s.weightRepo.GetBySubject(subjectID)
	if err != nil {
		return nil, err
	}

	if len(weights) == 0 {
		return nil, errors.New("no weights configured for this subject")
	}

	subject := weights[0].Subject

	weightDetails := []dto.WeightDetailDTO{}
	totalWeight := 0.0

	for _, w := range weights {
		weightDetails = append(weightDetails, dto.WeightDetailDTO{
			WeightID:     w.ID,
			CategoryID:   w.CategoryID,
			CategoryName: w.Category.Name,
			Weight:       w.Weight,
		})
		totalWeight += w.Weight
	}

	return &dto.WeightResponseDTO{
		SubjectID:   subject.ID,
		SubjectName: subject.Name,
		SubjectCode: subject.Code,
		Weights:     weightDetails,
		TotalWeight: totalWeight,
	}, nil
}

func (s *gradeService) ensureWeightSubjectInSchool(subjectID string, schoolID string) error {
	ok, err := s.weightRepo.SubjectBelongsToSchool(subjectID, schoolID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid assessment weight subject")
	}
	return nil
}

func (s *gradeService) ensureWeightCategoriesInSchool(categoryIDs []string, schoolID string) error {
	count, err := s.weightRepo.CountCategoriesInSchool(categoryIDs, schoolID)
	if err != nil {
		return err
	}
	if count != int64(len(categoryIDs)) {
		return fmt.Errorf("invalid assessment weight category")
	}
	return nil
}

func (s *gradeService) CalculateFinalGrade(studentID string, subjectID string) (*dto.GradeReportDTO, error) {
	weights, err := s.weightRepo.GetBySubject(subjectID)
	if err != nil {
		return nil, err
	}

	if len(weights) == 0 {
		return nil, errors.New("no weights configured for this subject")
	}

	assessments, err := s.gradeRepo.GetAssessmentsByStudentAndSubject(studentID, subjectID)
	if err != nil {
		return nil, err
	}

	categoryScores := make(map[string][]float64)
	for _, assessment := range assessments {
		categoryID := assessment.Submission.Assignment.CategoryID
		categoryScores[categoryID] = append(categoryScores[categoryID], assessment.Score)
	}

	breakdown := []dto.CategoryBreakdownDTO{}
	finalGrade := 0.0

	for _, weight := range weights {
		scores := categoryScores[weight.CategoryID]
		avgScore := calculateAverage(scores)
		weightedScore := avgScore * (weight.Weight / 100.0)
		finalGrade += weightedScore

		breakdown = append(breakdown, dto.CategoryBreakdownDTO{
			CategoryID:      weight.CategoryID,
			CategoryName:    weight.Category.Name,
			Weight:          weight.Weight,
			AverageScore:    avgScore,
			WeightedScore:   weightedScore,
			AssignmentCount: len(scores),
		})
	}

	user, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return nil, err
	}

	return &dto.GradeReportDTO{
		StudentID:   studentID,
		StudentName: user.FullName,
		SubjectID:   subjectID,
		SubjectName: weights[0].Subject.Name,
		Breakdown:   breakdown,
		FinalGrade:  finalGrade,
		LetterGrade: convertToLetterGrade(finalGrade),
	}, nil
}

func (s *gradeService) GetClassGradeReport(classID, subjectID, schoolID string) (*dto.ClassGradeReportDTO, error) {
	class, err := s.classRepo.GetByID(classID)
	if err != nil {
		return nil, err
	}
	if class.SchoolID != schoolID {
		return nil, fmt.Errorf("forbidden: class does not belong to active school")
	}

	subject, err := s.subjectRepo.GetByID(subjectID)
	if err != nil {
		return nil, err
	}
	if subject.SchoolID != schoolID {
		return nil, fmt.Errorf("forbidden: subject does not belong to active school")
	}

	students, err := s.gradeRepo.GetStudentsBySubjectClass(classID)
	if err != nil {
		return nil, err
	}

	studentGrades := []dto.StudentGradeSummaryDTO{}

	weights, err := s.weightRepo.GetBySubject(subjectID)
	if err == nil && len(weights) > 0 {
		studentIDs := make([]string, 0, len(students))
		for _, student := range students {
			studentIDs = append(studentIDs, student.ID)
		}

		assessments, err := s.gradeRepo.GetAssessmentsByStudentsAndSubject(studentIDs, subjectID)
		if err == nil {
			categoryScoresByStudent := make(map[string]map[string][]float64, len(students))
			for _, assessment := range assessments {
				studentID := assessment.Submission.UserID
				categoryID := assessment.Submission.Assignment.CategoryID
				if categoryScoresByStudent[studentID] == nil {
					categoryScoresByStudent[studentID] = make(map[string][]float64)
				}
				categoryScoresByStudent[studentID][categoryID] = append(categoryScoresByStudent[studentID][categoryID], assessment.Score)
			}

			for _, student := range students {
				categoryScores := categoryScoresByStudent[student.ID]

				finalGrade := 0.0
				for _, weight := range weights {
					avgScore := calculateAverage(categoryScores[weight.CategoryID])
					finalGrade += avgScore * (weight.Weight / 100.0)
				}

				// student.FullName/student.Email already come from
				// GetStudentsBySubjectClass above, so no per-student user
				// lookup is needed here.
				studentGrades = append(studentGrades, dto.StudentGradeSummaryDTO{
					StudentID:    student.ID,
					StudentName:  student.FullName,
					StudentEmail: student.Email,
					FinalGrade:   finalGrade,
					LetterGrade:  convertToLetterGrade(finalGrade),
				})
			}
		}
	}

	return &dto.ClassGradeReportDTO{
		Class: dto.ClassHeaderDTO{
			ID:    class.ID,
			Title: class.Title,
			Code:  class.Code,
		},
		Subject: dto.SubjectHeaderDTO{
			SubjectID:   subject.ID,
			SubjectName: subject.Name,
			SubjectCode: subject.Code,
		},
		Students: studentGrades,
	}, nil
}


func (s *gradeService) GetStudentGradeDetail(classID, subjectID, studentID, schoolID string) (*dto.StudentGradeDetailDTO, error) {
	classRow, err := s.gradeRepo.GetStudentGradebookClass(studentID, schoolID, classID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotEnrolledInClass
		}
		return nil, err
	}

	subject, err := s.subjectRepo.GetByID(subjectID)
	if err != nil {
		return nil, err
	}
	if subject.SchoolID != schoolID {
		return nil, fmt.Errorf("forbidden: subject does not belong to active school")
	}

	report, err := s.CalculateFinalGrade(studentID, subjectID)
	if err != nil {
		return nil, err
	}

	student, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return nil, err
	}

	rows, err := s.gradeRepo.GetStudentGradebookRows(studentID, schoolID, classID)
	if err != nil {
		return nil, err
	}

	assignments := []dto.MyGradebookAssignmentDTO{}
	for _, row := range rows {
		if row.SubjectID != subjectID || row.AssignmentID == nil {
			continue
		}

		status := "not_submitted"
		if row.SubmissionID != nil {
			status = "submitted"
		}
		if row.Score != nil {
			status = "graded"
		}

		assignments = append(assignments, dto.MyGradebookAssignmentDTO{
			AssignmentID:    *row.AssignmentID,
			AssignmentTitle: stringValue(row.AssignmentTitle),
			CategoryName:    stringValue(row.CategoryName),
			Deadline:        row.Deadline,
			Status:          status,
			SubmittedAt:     formatTimePointer(row.SubmittedAt),
			Score:           row.Score,
			Feedback:        row.Feedback,
			AssessedAt:      formatTimePointer(row.AssessedAt),
			AssessorName:    row.AssessorName,
		})
	}

	return &dto.StudentGradeDetailDTO{
		StudentID:    studentID,
		StudentName:  student.FullName,
		StudentEmail: student.Email,
		Class: dto.ClassHeaderDTO{
			ID:    classRow.ClassID,
			Title: classRow.ClassName,
			Code:  classRow.ClassCode,
		},
		Subject: dto.SubjectHeaderDTO{
			SubjectID:   subject.ID,
			SubjectName: subject.Name,
			SubjectCode: subject.Code,
		},
		FinalGrade:  report.FinalGrade,
		LetterGrade: report.LetterGrade,
		Breakdown:   report.Breakdown,
		Assignments: assignments,
	}, nil
}

func (s *gradeService) GetStudentReport(classID, studentID, schoolID string) (*dto.StudentReportDTO, error) {
	classRow, err := s.gradeRepo.GetStudentGradebookClass(studentID, schoolID, classID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotEnrolledInClass
		}
		return nil, err
	}

	student, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return nil, err
	}

	rows, err := s.gradeRepo.GetStudentGradebookRows(studentID, schoolID, classID)
	if err != nil {
		return nil, err
	}

	type subjectAccumulator struct {
		header      dto.SubjectHeaderDTO
		assignments []dto.MyGradebookAssignmentDTO
	}

	subjectOrder := make([]string, 0)
	bySubject := make(map[string]*subjectAccumulator)

	for _, row := range rows {
		acc, exists := bySubject[row.SubjectID]
		if !exists {
			acc = &subjectAccumulator{
				header: dto.SubjectHeaderDTO{
					SubjectID:   row.SubjectID,
					SubjectName: row.SubjectName,
					SubjectCode: row.SubjectCode,
				},
				assignments: []dto.MyGradebookAssignmentDTO{},
			}
			bySubject[row.SubjectID] = acc
			subjectOrder = append(subjectOrder, row.SubjectID)
		}

		if row.AssignmentID == nil {
			continue
		}

		status := "not_submitted"
		if row.SubmissionID != nil {
			status = "submitted"
		}
		if row.Score != nil {
			status = "graded"
		}

		acc.assignments = append(acc.assignments, dto.MyGradebookAssignmentDTO{
			AssignmentID:    *row.AssignmentID,
			AssignmentTitle: stringValue(row.AssignmentTitle),
			CategoryName:    stringValue(row.CategoryName),
			Deadline:        row.Deadline,
			Status:          status,
			SubmittedAt:     formatTimePointer(row.SubmittedAt),
			Score:           row.Score,
			Feedback:        row.Feedback,
			AssessedAt:      formatTimePointer(row.AssessedAt),
			AssessorName:    row.AssessorName,
		})
	}

	subjects := make([]dto.StudentReportSubjectDTO, 0, len(subjectOrder))
	totalFinalGrade := 0.0

	for _, subjectID := range subjectOrder {
		acc := bySubject[subjectID]

		report, err := s.CalculateFinalGrade(studentID, subjectID)
		if err != nil {
			continue
		}

		totalFinalGrade += report.FinalGrade
		subjects = append(subjects, dto.StudentReportSubjectDTO{
			Subject:     acc.header,
			FinalGrade:  report.FinalGrade,
			LetterGrade: report.LetterGrade,
			Breakdown:   report.Breakdown,
			Assignments: acc.assignments,
		})
	}

	averageFinalGrade := 0.0
	if len(subjects) > 0 {
		averageFinalGrade = totalFinalGrade / float64(len(subjects))
	}

	return &dto.StudentReportDTO{
		StudentID:    studentID,
		StudentName:  student.FullName,
		StudentEmail: student.Email,
		Class: dto.ClassHeaderDTO{
			ID:    classRow.ClassID,
			Title: classRow.ClassName,
			Code:  classRow.ClassCode,
		},
		Subjects: subjects,
		Summary: dto.StudentReportSummaryDTO{
			TotalSubjects:     len(subjects),
			AverageFinalGrade: averageFinalGrade,
		},
	}, nil
}

func (s *gradeService) GetMyGradebookByClass(userID string, schoolID string, classID string) (*dto.MyGradebookResponseDTO, error) {
	class, err := s.gradeRepo.GetStudentGradebookClass(userID, schoolID, classID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotEnrolledInClass
		}
		return nil, err
	}

	rows, err := s.gradeRepo.GetStudentGradebookRows(userID, schoolID, classID)
	if err != nil {
		return nil, err
	}

	response := &dto.MyGradebookResponseDTO{
		Class: dto.MyGradebookClassDTO{
			ClassID:   class.ClassID,
			ClassName: class.ClassName,
			ClassCode: class.ClassCode,
		},
		Subjects: []dto.MyGradebookSubjectDTO{},
		Summary:  dto.MyGradebookSummaryDTO{},
	}

	subjectIndexes := make(map[string]int)
	categoryScoresBySubject := make(map[string]map[string][]float64)

	for _, row := range rows {
		subjectIndex, exists := subjectIndexes[row.SubjectClassID]
		if !exists {
			response.Subjects = append(response.Subjects, dto.MyGradebookSubjectDTO{
				SubjectClassID: row.SubjectClassID,
				SubjectID:      row.SubjectID,
				SubjectName:    row.SubjectName,
				SubjectCode:    row.SubjectCode,
				Assignments:    []dto.MyGradebookAssignmentDTO{},
			})
			subjectIndex = len(response.Subjects) - 1
			subjectIndexes[row.SubjectClassID] = subjectIndex
			categoryScoresBySubject[row.SubjectClassID] = make(map[string][]float64)
		}

		if row.AssignmentID == nil {
			continue
		}

		status := "not_submitted"
		if row.SubmissionID != nil {
			status = "submitted"
			response.Subjects[subjectIndex].SubmittedCount++
			response.Summary.SubmittedAssignmentCount++
			response.Subjects[subjectIndex].PendingCount++
			response.Summary.PendingAssessmentCount++
		}
		if row.Score != nil {
			status = "graded"
			response.Subjects[subjectIndex].GradedCount++
			response.Summary.GradedAssignmentCount++
			response.Subjects[subjectIndex].PendingCount--
			response.Summary.PendingAssessmentCount--
			if row.CategoryID != nil {
				categoryScoresBySubject[row.SubjectClassID][*row.CategoryID] = append(categoryScoresBySubject[row.SubjectClassID][*row.CategoryID], *row.Score)
			}
		}

		response.Subjects[subjectIndex].Assignments = append(response.Subjects[subjectIndex].Assignments, dto.MyGradebookAssignmentDTO{
			AssignmentID:    *row.AssignmentID,
			AssignmentTitle: stringValue(row.AssignmentTitle),
			CategoryName:    stringValue(row.CategoryName),
			Deadline:        row.Deadline,
			Status:          status,
			SubmittedAt:     formatTimePointer(row.SubmittedAt),
			Score:           row.Score,
			Feedback:        row.Feedback,
			AssessedAt:      formatTimePointer(row.AssessedAt),
			AssessorName:    row.AssessorName,
		})
	}

	// weights are fetched once for all subjects in this gradebook instead of once per subject 
	subjectIDs := make([]string, 0, len(response.Subjects))
	for i := range response.Subjects {
		subjectIDs = append(subjectIDs, response.Subjects[i].SubjectID)
	}
	weightsBySubject, _ := s.weightRepo.GetBySubjects(subjectIDs)

	for i := range response.Subjects {
		subject := &response.Subjects[i]
		response.Summary.SubjectCount++
		finalGrade, letterGrade := s.calculateSubjectFinalGrade(weightsBySubject[subject.SubjectID], categoryScoresBySubject[subject.SubjectClassID])
		subject.FinalGrade = finalGrade
		subject.LetterGrade = letterGrade
	}

	return response, nil
}

func calculateAverage(scores []float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

func (s *gradeService) calculateSubjectFinalGrade(weights []*domain.AssessmentWeight, categoryScores map[string][]float64) (*float64, *string) {
	if len(categoryScores) == 0 {
		return nil, nil
	}

	if len(weights) == 0 {
		return nil, nil
	}

	finalGrade := 0.0
	hasWeightedScore := false
	for _, weight := range weights {
		scores := categoryScores[weight.CategoryID]
		if len(scores) == 0 {
			continue
		}
		finalGrade += calculateAverage(scores) * (weight.Weight / 100.0)
		hasWeightedScore = true
	}

	if !hasWeightedScore {
		return nil, nil
	}

	letterGrade := convertToLetterGrade(finalGrade)
	return &finalGrade, &letterGrade
}

func formatTimePointer(value *time.Time) *string {
	return formatAPITimePtr(value)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func convertToLetterGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "E"
	}
}
