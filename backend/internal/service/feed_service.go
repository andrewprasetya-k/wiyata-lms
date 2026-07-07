package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/events"
	"backend/internal/repository"
	"errors"
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"
)

type FeedService interface {
	Create(feed *domain.Feed, userID string, roles []string) error
	GetByClass(classID string, schoolID string, userID string, roles []string, page int, limit int) ([]*domain.Feed, int64, error)
	GetByID(id string, schoolID string, userID string, roles []string) (*domain.Feed, error)
	Update(id string, schoolID string, userID string, roles []string, content *string) error
	Delete(id string, schoolID string, userID string, roles []string) error
}

type feedService struct {
	repo             repository.FeedRepository
	attService       AttachmentService
	notifService     NotificationService
	enrRepo          repository.EnrollmentRepository
	classRepo        repository.ClassRepository
	subjectClassRepo repository.SubjectClassRepository
	broadcaster      events.SidebarBroadcaster
}

func NewFeedService(repo repository.FeedRepository, attService AttachmentService, notifService NotificationService, enrRepo repository.EnrollmentRepository, classRepo repository.ClassRepository, subjectClassRepo repository.SubjectClassRepository, broadcaster events.SidebarBroadcaster) FeedService {
	return &feedService{
		repo:             repo,
		attService:       attService,
		notifService:     notifService,
		enrRepo:          enrRepo,
		classRepo:        classRepo,
		subjectClassRepo: subjectClassRepo,
		broadcaster:      broadcaster,
	}
}

func (s *feedService) Create(feed *domain.Feed, userID string, roles []string) error {
	feed.Content = strings.TrimSpace(feed.Content)
	if feed.Content == "" {
		return fmt.Errorf("feed content is required")
	}

	classSchoolID, err := s.classRepo.GetSchoolIDByClass(feed.ClassID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return err
	}
	if classSchoolID != feed.SchoolID {
		return fmt.Errorf("forbidden: class does not belong to active school")
	}

	if !hasFeedRole(roles, "admin") && !hasFeedRole(roles, "teacher") {
		return fmt.Errorf("forbidden: insufficient feed permission")
	}

	if hasFeedRole(roles, "teacher") && !hasFeedRole(roles, "admin") {
		if err := s.ensureTeacherCanAccessClass(userID, feed.SchoolID, feed.ClassID); err != nil {
			return err
		}
	}

	if err := s.repo.Create(feed); err != nil {
		return err
	}

	// Best-effort: notify all class members except the creator
	if userIDs, err := s.enrRepo.GetMemberUserIDsByClass(feed.ClassID); err == nil {
		className := ""
		if class, classErr := s.classRepo.GetByID(feed.ClassID); classErr == nil {
			className = strings.TrimSpace(class.Title)
			if className == "" {
				className = strings.TrimSpace(class.Code)
			}
		}

		message := feedNotificationMessage(className, feed.Content)
		for _, uid := range userIDs {
			if uid == feed.CreatedBy {
				continue
			}
			_ = s.notifService.Create(&dto.CreateNotificationDTO{
				UserID:    uid,
				Type:      domain.NotifFeedPosted,
				Title:     "Pengumuman kelas baru",
				Message:   message,
				Link:      "/student/feed",
				RelatedID: feed.ID,
			})
		}
	}

	return nil
}

func (s *feedService) GetByClass(classID string, schoolID string, userID string, roles []string, page int, limit int) ([]*domain.Feed, int64, error) {
	if err := s.ensureCanReadClassFeed(userID, schoolID, classID, roles); err != nil {
		return nil, 0, err
	}

	feeds, total, err := s.repo.GetByClassInSchool(classID, schoolID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	sourceIDs := make([]string, 0, len(feeds))
	for _, f := range feeds {
		sourceIDs = append(sourceIDs, f.ID)
	}
	attachmentsBySource, err := s.attService.GetBySources(string(domain.SourceFeed), sourceIDs)
	if err != nil {
		return nil, 0, err
	}

	for _, f := range feeds {
		atts := attachmentsBySource[f.ID]
		f.Attachments = make([]domain.Attachment, 0, len(atts))
		for _, a := range atts {
			f.Attachments = append(f.Attachments, *a)
		}
	}

	return feeds, total, nil
}

func (s *feedService) GetByID(id string, schoolID string, userID string, roles []string) (*domain.Feed, error) {
	feed, err := s.repo.GetByIDInSchool(id, schoolID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCanReadClassFeed(userID, schoolID, feed.ClassID, roles); err != nil {
		return nil, err
	}

	atts, _ := s.attService.GetBySource(string(domain.SourceFeed), id)
	for _, a := range atts {
		feed.Attachments = append(feed.Attachments, *a)
	}

	return feed, nil
}

func (s *feedService) Update(id string, schoolID string, userID string, roles []string, content *string) error {
	feed, err := s.repo.GetByIDInSchool(id, schoolID)
	if err != nil {
		return err
	}
	if err := s.ensureCanMutateFeed(feed, schoolID, userID, roles); err != nil {
		return err
	}
	if content != nil {
		feed.Content = strings.TrimSpace(*content)
	}
	if feed.Content == "" {
		return fmt.Errorf("feed content is required")
	}

	return s.repo.UpdateInSchool(feed, schoolID)
}

func (s *feedService) Delete(id string, schoolID string, userID string, roles []string) error {
	feed, err := s.repo.GetByIDInSchool(id, schoolID)
	if err != nil {
		return err
	}
	if err := s.ensureCanMutateFeed(feed, schoolID, userID, roles); err != nil {
		return err
	}
	if err := s.repo.DeleteInSchool(id, schoolID); err != nil {
		return err
	}

	if s.broadcaster != nil {
		if userIDs, membersErr := s.enrRepo.GetMemberUserIDsByClass(feed.ClassID); membersErr == nil {
			s.broadcaster.BroadcastToUsers(schoolID, userIDs, events.SidebarEvent{
				Type:     events.SidebarEventTypeFeedChanged,
				SchoolID: schoolID,
			})
		}
	}

	return nil
}

func (s *feedService) ensureCanReadClassFeed(userID string, schoolID string, classID string, roles []string) error {
	classSchoolID, err := s.classRepo.GetSchoolIDByClass(classID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return err
	}
	if classSchoolID != schoolID {
		return fmt.Errorf("forbidden: class does not belong to active school")
	}

	if hasFeedRole(roles, "admin") {
		return nil
	}
	if hasFeedRole(roles, "teacher") {
		if err := s.ensureTeacherCanAccessClass(userID, schoolID, classID); err == nil {
			return nil
		}
	}
	if hasFeedRole(roles, "student") {
		ok, err := s.enrRepo.UserEnrolledInClassAsRole(userID, schoolID, classID, "student")
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return fmt.Errorf("forbidden: feed class access denied")
}

func (s *feedService) ensureCanMutateFeed(feed *domain.Feed, schoolID string, userID string, roles []string) error {
	if feed.SchoolID != schoolID {
		return fmt.Errorf("forbidden: feed does not belong to active school")
	}
	if hasFeedRole(roles, "admin") {
		return nil
	}
	if hasFeedRole(roles, "teacher") && feed.CreatedBy == userID {
		return s.ensureTeacherCanAccessClass(userID, schoolID, feed.ClassID)
	}
	return fmt.Errorf("forbidden: feed mutation is not allowed")
}

func (s *feedService) ensureTeacherCanAccessClass(userID string, schoolID string, classID string) error {
	canPost, err := s.subjectClassRepo.UserTeachesClass(userID, schoolID, classID)
	if err != nil {
		return err
	}
	if !canPost {
		return fmt.Errorf("forbidden: teacher does not teach this class")
	}
	return nil
}

func hasFeedRole(roles []string, role string) bool {
	return slices.Contains(roles, role)
}

func feedNotificationMessage(className string, content string) string {
	preview := strings.Join(strings.Fields(content), " ")
	previewRunes := []rune(preview)
	if len(previewRunes) > 120 {
		preview = string(previewRunes[:117]) + "..."
	}
	if className == "" {
		return preview
	}
	if preview == "" {
		return className
	}
	return fmt.Sprintf("%s • %s", className, preview)
}
