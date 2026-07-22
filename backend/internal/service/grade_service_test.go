package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"strings"
	"testing"
)

// ─── AssessmentWeightRepository stub ─────────────────────────────────────────

type gradeWeightRepoStub struct {
	subjectInSchool    bool
	subjectErr         error
	categoriesAllValid bool
	categoriesErr      error
	replacedWeights    []*domain.AssessmentWeight
	replaceErr         error
}

func (r *gradeWeightRepoStub) SubjectBelongsToSchool(_, _ string) (bool, error) {
	return r.subjectInSchool, r.subjectErr
}
func (r *gradeWeightRepoStub) CountCategoriesInSchool(ids []string, _ string) (int64, error) {
	if r.categoriesErr != nil {
		return 0, r.categoriesErr
	}
	if r.categoriesAllValid {
		return int64(len(ids)), nil
	}
	return 0, nil
}
func (r *gradeWeightRepoStub) ReplaceBySubject(_ string, weights []*domain.AssessmentWeight) error {
	r.replacedWeights = weights
	return r.replaceErr
}
func (r *gradeWeightRepoStub) Create(_ *domain.AssessmentWeight) error { return nil }
func (r *gradeWeightRepoStub) GetBySubject(_ string) ([]*domain.AssessmentWeight, error) {
	return nil, nil
}
func (r *gradeWeightRepoStub) GetBySubjects(_ []string) (map[string][]*domain.AssessmentWeight, error) {
	return nil, nil
}
func (r *gradeWeightRepoStub) DeleteBySubject(_ string) error                    { return nil }
func (r *gradeWeightRepoStub) GetTotalWeightBySubject(_ string) (float64, error) { return 0, nil }

// ─── Minimal no-op stubs for GradeService's other repos ──────────────────────

type gradeTestGradeRepoStub struct{}

func (r *gradeTestGradeRepoStub) GetAssessmentsByStudentAndSubject(_, _ string) ([]*domain.Assessment, error) {
	return nil, nil
}
func (r *gradeTestGradeRepoStub) GetAssessmentsByStudentsAndSubject(_ []string, _ string) ([]*domain.Assessment, error) {
	return nil, nil
}
func (r *gradeTestGradeRepoStub) GetStudentsBySubjectClass(_ string) ([]*domain.User, error) {
	return nil, nil
}
func (r *gradeTestGradeRepoStub) GetStudentGradebookClass(_, _, _ string) (*dto.StudentGradebookClassRow, error) {
	return nil, nil
}
func (r *gradeTestGradeRepoStub) GetStudentGradebookRows(_, _, _ string) ([]dto.StudentGradebookRow, error) {
	return nil, nil
}

type gradeTestSubjectRepoStub struct{}

func (r *gradeTestSubjectRepoStub) Create(_ *domain.Subject) error { return nil }
func (r *gradeTestSubjectRepoStub) FindAll(_ string, _ string, _, _ int) ([]*domain.Subject, int64, error) {
	return nil, 0, nil
}
func (r *gradeTestSubjectRepoStub) GetBySchool(_ string) ([]*domain.Subject, error) { return nil, nil }
func (r *gradeTestSubjectRepoStub) GetByID(_ string) (*domain.Subject, error)       { return nil, nil }
func (r *gradeTestSubjectRepoStub) GetByCode(_, _ string) (*domain.Subject, error)  { return nil, nil }
func (r *gradeTestSubjectRepoStub) Update(_ *domain.Subject) error                  { return nil }
func (r *gradeTestSubjectRepoStub) Delete(_ string) error                           { return nil }
func (r *gradeTestSubjectRepoStub) CheckDuplicateCode(_, _, _ string) (bool, error) {
	return false, nil
}
func (r *gradeTestSubjectRepoStub) CountSubjectClassesBySubject(_ string) (int64, error) {
	return 0, nil
}

type gradeTestClassRepoStub struct{}

func (r *gradeTestClassRepoStub) Create(_ *domain.Class) error { return nil }
func (r *gradeTestClassRepoStub) FindAll(_, _, _ string, _, _ int) ([]*domain.Class, int64, error) {
	return nil, 0, nil
}
func (r *gradeTestClassRepoStub) GetByID(_ string) (*domain.Class, error)            { return nil, nil }
func (r *gradeTestClassRepoStub) Update(_ *domain.Class) error                       { return nil }
func (r *gradeTestClassRepoStub) Delete(_ string) error                              { return nil }
func (r *gradeTestClassRepoStub) CountEnrollmentsByClass(_ string) (int64, error)    { return 0, nil }
func (r *gradeTestClassRepoStub) CountSubjectClassesByClass(_ string) (int64, error) { return 0, nil }
func (r *gradeTestClassRepoStub) CheckDuplicateCode(_, _, _, _ string) (bool, error) {
	return false, nil
}
func (r *gradeTestClassRepoStub) GetSchoolIDByClass(_ string) (string, error) { return "", nil }

type gradeTestUserRepoStub struct{}

func (r *gradeTestUserRepoStub) Create(_ *domain.User) error { return nil }
func (r *gradeTestUserRepoStub) FindAll(_ string, _, _ int) ([]*domain.User, int64, error) {
	return nil, 0, nil
}
func (r *gradeTestUserRepoStub) GetByID(_ string) (*domain.User, error)     { return nil, nil }
func (r *gradeTestUserRepoStub) GetByEmail(_ string) (*domain.User, error)  { return nil, nil }
func (r *gradeTestUserRepoStub) Update(_ *domain.User) error                { return nil }
func (r *gradeTestUserRepoStub) Delete(_ string) error                      { return nil }
func (r *gradeTestUserRepoStub) CheckEmailExists(_, _ string) (bool, error) { return false, nil }

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newGradeWeightsTestService(stub *gradeWeightRepoStub) GradeService {
	return NewGradeService(
		stub,
		&gradeTestGradeRepoStub{},
		&gradeTestSubjectRepoStub{},
		&gradeTestClassRepoStub{},
		&gradeTestUserRepoStub{},
		nil,
	)
}

func validWeightsStub() *gradeWeightRepoStub {
	return &gradeWeightRepoStub{subjectInSchool: true, categoriesAllValid: true}
}

func wfloatPtr(f float64) *float64 { return &f }

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestGradeConfigureWeightsRejectsEmptyList(t *testing.T) {
	svc := newGradeWeightsTestService(validWeightsStub())
	err := svc.ConfigureWeights(domain.ActorContext{}, &dto.ConfigureWeightsDTO{SubjectID: "sub-1", Weights: nil}, "school-1")
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestGradeConfigureWeightsRejectsNilWeightPointer(t *testing.T) {
	svc := newGradeWeightsTestService(validWeightsStub())
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights:   []dto.WeightItemDTO{{CategoryID: "cat-1", Weight: nil}},
	}
	err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1")
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestGradeConfigureWeightsRejectsWeightOutOfRange(t *testing.T) {
	cases := []struct {
		name   string
		weight float64
	}{
		{"negative", -1},
		{"above100", 101},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newGradeWeightsTestService(validWeightsStub())
			req := &dto.ConfigureWeightsDTO{
				SubjectID: "sub-1",
				Weights:   []dto.WeightItemDTO{{CategoryID: "cat-1", Weight: wfloatPtr(tc.weight)}},
			}
			err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1")
			if err == nil || !strings.Contains(err.Error(), "between 0 and 100") {
				t.Fatalf("expected range error for weight %v, got %v", tc.weight, err)
			}
		})
	}
}

func TestGradeConfigureWeightsRejectsTotalNotHundred(t *testing.T) {
	svc := newGradeWeightsTestService(validWeightsStub())
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights: []dto.WeightItemDTO{
			{CategoryID: "cat-1", Weight: wfloatPtr(60)},
			{CategoryID: "cat-2", Weight: wfloatPtr(30)},
		},
	}
	err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1")
	if err == nil || !strings.Contains(err.Error(), "total weight must be 100") {
		t.Fatalf("expected total weight error, got %v", err)
	}
}

func TestGradeConfigureWeightsRejectsDuplicateCategory(t *testing.T) {
	svc := newGradeWeightsTestService(validWeightsStub())
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights: []dto.WeightItemDTO{
			{CategoryID: "cat-1", Weight: wfloatPtr(50)},
			{CategoryID: "cat-1", Weight: wfloatPtr(50)},
		},
	}
	err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1")
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate category error, got %v", err)
	}
}

func TestGradeConfigureWeightsRejectsSubjectFromOtherSchool(t *testing.T) {
	stub := &gradeWeightRepoStub{subjectInSchool: false, categoriesAllValid: true}
	svc := newGradeWeightsTestService(stub)
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights:   []dto.WeightItemDTO{{CategoryID: "cat-1", Weight: wfloatPtr(100)}},
	}
	err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1")
	if err == nil {
		t.Fatal("expected error for subject from another school, got nil")
	}
}

func TestGradeConfigureWeightsRejectsCategoryFromOtherSchool(t *testing.T) {
	stub := &gradeWeightRepoStub{subjectInSchool: true, categoriesAllValid: false}
	svc := newGradeWeightsTestService(stub)
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights:   []dto.WeightItemDTO{{CategoryID: "cat-1", Weight: wfloatPtr(100)}},
	}
	err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("expected invalid category error, got %v", err)
	}
}

func TestGradeConfigureWeightsCallsReplaceBySubjectNotCreate(t *testing.T) {
	stub := validWeightsStub()
	svc := newGradeWeightsTestService(stub)
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights: []dto.WeightItemDTO{
			{CategoryID: "cat-1", Weight: wfloatPtr(60)},
			{CategoryID: "cat-2", Weight: wfloatPtr(40)},
		},
	}
	if err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.replacedWeights == nil {
		t.Fatal("expected ReplaceBySubject to be called, got nil")
	}
	if len(stub.replacedWeights) != 2 {
		t.Fatalf("expected 2 weights passed to ReplaceBySubject, got %d", len(stub.replacedWeights))
	}
}

func TestGradeConfigureWeightsAcceptsExactHundred(t *testing.T) {
	stub := validWeightsStub()
	svc := newGradeWeightsTestService(stub)
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights: []dto.WeightItemDTO{
			{CategoryID: "cat-1", Weight: wfloatPtr(40)},
			{CategoryID: "cat-2", Weight: wfloatPtr(30)},
			{CategoryID: "cat-3", Weight: wfloatPtr(30)},
		},
	}
	if err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1"); err != nil {
		t.Fatalf("unexpected error for valid weights summing to 100: %v", err)
	}
}

func TestGradeConfigureWeightsToleratesFloatingPointTotal(t *testing.T) {
	stub := validWeightsStub()
	svc := newGradeWeightsTestService(stub)
	// 33.33 + 33.33 + 33.34 = 100.00 exactly — valid despite float arithmetic
	req := &dto.ConfigureWeightsDTO{
		SubjectID: "sub-1",
		Weights: []dto.WeightItemDTO{
			{CategoryID: "cat-1", Weight: wfloatPtr(33.33)},
			{CategoryID: "cat-2", Weight: wfloatPtr(33.33)},
			{CategoryID: "cat-3", Weight: wfloatPtr(33.34)},
		},
	}
	if err := svc.ConfigureWeights(domain.ActorContext{}, req, "school-1"); err != nil {
		t.Fatalf("unexpected error for 33.33+33.33+33.34 total: %v", err)
	}
}
