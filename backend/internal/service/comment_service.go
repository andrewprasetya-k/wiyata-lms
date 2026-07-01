package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"backend/internal/utils"
	"fmt"
	"slices"
	"strings"
)

type CommentService interface {
	Create(comment *domain.Comment, schoolID string, userID string, roles []string) error
	GetBySource(sourceType string, sourceID string, schoolID string, userID string, roles []string) ([]*domain.Comment, error)
	GetByID(id string, schoolID string, userID string, roles []string) (*domain.Comment, error)
	Update(id string, schoolID string, userID string, roles []string, content *string) error
	Delete(id string, schoolID string, userID string, roles []string) error
	CountBySource(sourceType string, sourceID string, schoolID string) (int, error)
}

type commentService struct {
	repo             repository.CommentRepository
	contentOwnerRepo repository.ContentOwnerRepository
	notifService     NotificationService
	feedRepo         repository.FeedRepository
	enrRepo          repository.EnrollmentRepository
	subjectClassRepo repository.SubjectClassRepository
}

func NewCommentService(repo repository.CommentRepository, contentOwnerRepo repository.ContentOwnerRepository, notifService NotificationService, feedRepo repository.FeedRepository, enrRepo repository.EnrollmentRepository, subjectClassRepo repository.SubjectClassRepository) CommentService {
	return &commentService{
		repo:             repo,
		contentOwnerRepo: contentOwnerRepo,
		notifService:     notifService,
		feedRepo:         feedRepo,
		enrRepo:          enrRepo,
		subjectClassRepo: subjectClassRepo,
	}
}

func (s *commentService) Create(comment *domain.Comment, schoolID string, userID string, roles []string) error {
	comment.SchoolID = schoolID
	comment.UserID = userID
	comment.Content = strings.TrimSpace(comment.Content)
	if comment.Content == "" {
		return fmt.Errorf("comment content is required")
	}
	if err := s.ensureCanAccessSource(comment.SourceType, comment.SourceID, schoolID, userID, roles); err != nil {
		return err
	}

	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = utils.NowJakarta()
	}

	if err := s.repo.Create(comment); err != nil {
		return err
	}

	// Best-effort: notify content owner, skip if self-comment
	if ownerID, err := s.contentOwnerRepo.GetOwnerUserID(comment.SourceType, comment.SourceID); err == nil && ownerID != "" && ownerID != comment.UserID {
		_ = s.notifService.Create(&dto.CreateNotificationDTO{
			UserID:    ownerID,
			Type:      domain.NotifCommentAdded,
			Title:     "New Comment",
			Message:   "Someone commented on your content.",
			RelatedID: comment.SourceID,
		})
	}

	return nil
}

func (s *commentService) GetBySource(sourceType string, sourceID string, schoolID string, userID string, roles []string) ([]*domain.Comment, error) {
	source := domain.SourceType(sourceType)
	if err := s.ensureCanAccessSource(source, sourceID, schoolID, userID, roles); err != nil {
		return nil, err
	}
	comments, err := s.repo.GetBySourceInSchool(source, sourceID, schoolID)
	if err != nil {
		return nil, err
	}
	if comments == nil {
		return []*domain.Comment{}, nil
	}
	return comments, nil
}

func (s *commentService) GetByID(id string, schoolID string, userID string, roles []string) (*domain.Comment, error) {
	comment, err := s.repo.GetByIDInSchool(id, schoolID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCanAccessSource(comment.SourceType, comment.SourceID, schoolID, userID, roles); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *commentService) Update(id string, schoolID string, userID string, roles []string, content *string) error {
	comment, err := s.repo.GetByIDInSchool(id, schoolID)
	if err != nil {
		return err
	}
	if comment.UserID != userID {
		return fmt.Errorf("forbidden: comment update is not allowed")
	}
	if err := s.ensureCanAccessSource(comment.SourceType, comment.SourceID, schoolID, userID, roles); err != nil {
		return err
	}
	if content != nil {
		comment.Content = strings.TrimSpace(*content)
	}
	if comment.Content == "" {
		return fmt.Errorf("comment content is required")
	}
	return s.repo.UpdateInSchool(comment, schoolID)
}

func (s *commentService) Delete(id string, schoolID string, userID string, roles []string) error {
	comment, err := s.repo.GetByIDInSchool(id, schoolID)
	if err != nil {
		return err
	}
	if comment.SourceType != domain.SourceFeed {
		return fmt.Errorf("unsupported comment source")
	}
	if !hasCommentRole(roles, "admin") && comment.UserID != userID {
		return fmt.Errorf("forbidden: comment delete is not allowed")
	}
	if !hasCommentRole(roles, "admin") {
		if err := s.ensureCanAccessSource(comment.SourceType, comment.SourceID, schoolID, userID, roles); err != nil {
			return err
		}
	}
	return s.repo.DeleteInSchool(id, schoolID)
}

func (s *commentService) CountBySource(sourceType string, sourceID string, schoolID string) (int, error) {
	source := domain.SourceType(sourceType)
	if source != domain.SourceFeed {
		return 0, nil
	}
	return s.repo.CountBySourceInSchool(source, sourceID, schoolID)
}

func (s *commentService) ensureCanAccessSource(sourceType domain.SourceType, sourceID string, schoolID string, userID string, roles []string) error {
	if sourceType != domain.SourceFeed {
		return fmt.Errorf("unsupported comment source")
	}

	feed, err := s.feedRepo.GetByIDInSchool(sourceID, schoolID)
	if err != nil {
		return err
	}

	if hasCommentRole(roles, "admin") {
		return nil
	}
	if hasCommentRole(roles, "teacher") {
		ok, err := s.subjectClassRepo.UserTeachesClass(userID, schoolID, feed.ClassID)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	if hasCommentRole(roles, "student") {
		ok, err := s.enrRepo.UserEnrolledInClassAsRole(userID, schoolID, feed.ClassID, "student")
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}

	return fmt.Errorf("forbidden: comment source access denied")
}

func hasCommentRole(roles []string, role string) bool {
	return slices.Contains(roles, role)
}
