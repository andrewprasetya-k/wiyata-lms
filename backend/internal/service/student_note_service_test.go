package service

import (
	"backend/internal/domain"
	"errors"
	"strings"
	"testing"
)

type studentNoteRepositoryStub struct {
	note      *domain.StudentNote
	savedNote *domain.StudentNote
	deleted   bool
}

func (r *studentNoteRepositoryStub) GetByUserMaterialInSchool(string, string, string) (*domain.StudentNote, error) {
	return r.note, nil
}

func (r *studentNoteRepositoryStub) Upsert(note *domain.StudentNote) (*domain.StudentNote, error) {
	r.savedNote = note
	return note, nil
}

func (r *studentNoteRepositoryStub) DeleteByUserMaterialInSchool(string, string, string) error {
	r.deleted = true
	return nil
}

type studentNoteMaterialRepositoryStub struct {
	material *domain.Material
	err      error
}

func (r *studentNoteMaterialRepositoryStub) GetByID(string) (*domain.Material, error) {
	return r.material, r.err
}

type studentNoteSubjectClassRepositoryStub struct {
	allowed bool
	err     error
}

func (r *studentNoteSubjectClassRepositoryStub) UserEnrolledInSubjectClassAsRole(string, string, string, string) (bool, error) {
	return r.allowed, r.err
}

func newStudentNoteTestService(noteRepo *studentNoteRepositoryStub, allowed bool) StudentNoteService {
	return NewStudentNoteService(
		noteRepo,
		&studentNoteMaterialRepositoryStub{
			material: &domain.Material{
				ID:             "material-1",
				SchoolID:       "school-1",
				SubjectClassID: "subject-class-1",
			},
		},
		&studentNoteSubjectClassRepositoryStub{allowed: allowed},
	)
}

func TestStudentNoteGetReturnsNilWhenNoteDoesNotExist(t *testing.T) {
	service := newStudentNoteTestService(&studentNoteRepositoryStub{}, true)

	note, err := service.GetMaterialNote("material-1", "school-1", "user-1")
	if err != nil {
		t.Fatalf("GetMaterialNote returned error: %v", err)
	}
	if note != nil {
		t.Fatalf("expected nil note, got %#v", note)
	}
}

func TestStudentNoteSaveTrimsAndPersistsContent(t *testing.T) {
	repo := &studentNoteRepositoryStub{}
	service := newStudentNoteTestService(repo, true)

	_, err := service.SaveMaterialNote("material-1", "school-1", "user-1", "  Ringkasan materi  ")
	if err != nil {
		t.Fatalf("SaveMaterialNote returned error: %v", err)
	}
	if repo.savedNote == nil || repo.savedNote.Content != "Ringkasan materi" {
		t.Fatalf("expected trimmed note content, got %#v", repo.savedNote)
	}
}

func TestStudentNoteSaveRejectsInvalidContent(t *testing.T) {
	service := newStudentNoteTestService(&studentNoteRepositoryStub{}, true)

	if _, err := service.SaveMaterialNote("material-1", "school-1", "user-1", "   "); !errors.Is(err, ErrStudentNoteContentRequired) {
		t.Fatalf("expected required content error, got %v", err)
	}
	if _, err := service.SaveMaterialNote("material-1", "school-1", "user-1", strings.Repeat("a", studentNoteMaxLength+1)); !errors.Is(err, ErrStudentNoteContentTooLong) {
		t.Fatalf("expected content length error, got %v", err)
	}
}

func TestStudentNoteRequiresActiveMaterialAccess(t *testing.T) {
	service := newStudentNoteTestService(&studentNoteRepositoryStub{}, false)

	_, err := service.GetMaterialNote("material-1", "school-1", "user-1")
	if err == nil || !strings.Contains(err.Error(), "forbidden:") {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestStudentNoteRejectsMaterialFromAnotherSchool(t *testing.T) {
	service := newStudentNoteTestService(&studentNoteRepositoryStub{}, true)

	_, err := service.GetMaterialNote("material-1", "school-2", "user-1")
	if err == nil || !strings.Contains(err.Error(), "forbidden:") {
		t.Fatalf("expected forbidden school error, got %v", err)
	}
}

func TestStudentNoteDeleteUsesOwnerScopedRepository(t *testing.T) {
	repo := &studentNoteRepositoryStub{}
	service := newStudentNoteTestService(repo, true)

	if err := service.DeleteMaterialNote("material-1", "school-1", "user-1"); err != nil {
		t.Fatalf("DeleteMaterialNote returned error: %v", err)
	}
	if !repo.deleted {
		t.Fatal("expected owner-scoped hard delete")
	}
}
