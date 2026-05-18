package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"errors"
	"fmt"
)

type GradeService interface {
	ConfigureWeights(req *dto.ConfigureWeightsDTO) error
	GetWeightsBySubject(subjectID string) (*dto.WeightResponseDTO, error)
	CalculateFinalGrade(studentID string, subjectID string) (*dto.GradeReportDTO, error)
	GetClassGradeReport(classID, subjectID string) (*dto.ClassGradeReportDTO, error)
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

func (s *gradeService) ConfigureWeights(req *dto.ConfigureWeightsDTO) error {
	totalWeight := 0.0
	for _, w := range req.Weights {
		totalWeight += w.Weight
	}

	if totalWeight != 100.0 {
		return fmt.Errorf("total weight must be 100, got %.2f", totalWeight)
	}

	if err := s.weightRepo.DeleteBySubject(req.SubjectID); err != nil {
		return err
	}

	for _, w := range req.Weights {
		weight := &domain.AssessmentWeight{
			SubjectID:  req.SubjectID,
			CategoryID: w.CategoryID,
			Weight:     w.Weight,
		}
		if err := s.weightRepo.Create(weight); err != nil {
			return err
		}
	}

	return nil
}

func (s *gradeService) GetWeightsBySubject(subjectID string) (*dto.WeightResponseDTO, error) {
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

func (s *gradeService) GetClassGradeReport(classID, subjectID string) (*dto.ClassGradeReportDTO, error) {
	class, err := s.classRepo.GetByID(classID)
	if err != nil {
		return nil, err
	}

	subject, err := s.subjectRepo.GetByID(subjectID)
	if err != nil {
		return nil, err
	}

	students, err := s.gradeRepo.GetStudentsBySubjectClass(classID)
	if err != nil {
		return nil, err
	}

	studentGrades := []dto.StudentGradeSummaryDTO{}

	for _, student := range students {
		grade, err := s.CalculateFinalGrade(student.ID, subjectID)
		if err != nil {
			continue
		}

		studentGrades = append(studentGrades, dto.StudentGradeSummaryDTO{
			StudentID:    student.ID,
			StudentName:  student.FullName,
			StudentEmail: student.Email,
			FinalGrade:   grade.FinalGrade,
			LetterGrade:  grade.LetterGrade,
		})
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
