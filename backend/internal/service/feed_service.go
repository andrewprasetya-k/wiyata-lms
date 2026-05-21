package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/repository"
	"fmt"
)

type FeedService interface {
	Create(feed *domain.Feed, mediaIDs []string, userID string, userRole string) error
	GetByClass(classID string, page int, limit int) ([]*domain.Feed, int64, error)
	GetByID(id string) (*domain.Feed, error)
	Update(id string, feed *domain.Feed, mediaIDs []string) error
	Delete(id string) error
}

type feedService struct {
	repo             repository.FeedRepository
	attService       AttachmentService
	notifService     NotificationService
	enrRepo          repository.EnrollmentRepository
	classRepo        repository.ClassRepository
	subjectClassRepo repository.SubjectClassRepository
}

func NewFeedService(repo repository.FeedRepository, attService AttachmentService, notifService NotificationService, enrRepo repository.EnrollmentRepository, classRepo repository.ClassRepository, subjectClassRepo repository.SubjectClassRepository) FeedService {
	return &feedService{
		repo:             repo,
		attService:       attService,
		notifService:     notifService,
		enrRepo:          enrRepo,
		classRepo:        classRepo,
		subjectClassRepo: subjectClassRepo,
	}
}

func (s *feedService) Create(feed *domain.Feed, mediaIDs []string, userID string, userRole string) error {

	classSchoolID, err := s.classRepo.GetSchoolIDByClass(feed.ClassID)
	if err != nil {
		return fmt.Errorf("class not found")
	}
	if classSchoolID != feed.SchoolID {
		return fmt.Errorf("class does not belong to this school")
	}

	if userRole == "teacher" {
		canPost, err := s.subjectClassRepo.TeacherTeachesInClass(userID, feed.ClassID)
		if err != nil {
			return fmt.Errorf("failed to verify teacher authorization")
		}
		if !canPost {
			return fmt.Errorf("teacher does not teach any subject in this class")
		}
	}

	if err := s.repo.Create(feed); err != nil {
		return err
	}

	for _, mID := range mediaIDs {
		att := &domain.Attachment{
			SchoolID:   feed.SchoolID,
			SourceID:   feed.ID,
			SourceType: domain.SourceFeed,
			MediaID:    mID,
		}
		s.attService.Link(att)
	}

	// Best-effort: notify all class members except the creator
	if userIDs, err := s.enrRepo.GetMemberUserIDsByClass(feed.ClassID); err == nil {
		for _, uid := range userIDs {
			if uid == feed.CreatedBy {
				continue
			}
			_ = s.notifService.Create(&dto.CreateNotificationDTO{
				UserID:    uid,
				Type:      domain.NotifFeedPosted,
				Title:     "New Announcement",
				Message:   "A new post has been made in your class.",
				RelatedID: feed.ID,
			})
		}
	}

	return nil
}

func (s *feedService) GetByClass(classID string, page int, limit int) ([]*domain.Feed, int64, error) {
	feeds, total, err := s.repo.GetByClass(classID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, f := range feeds {
		atts, _ := s.attService.GetBySource(string(domain.SourceFeed), f.ID)
		for _, a := range atts {
			f.Attachments = append(f.Attachments, *a)
		}
	}

	return feeds, total, nil
}

func (s *feedService) GetByID(id string) (*domain.Feed, error) {
	feed, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	atts, _ := s.attService.GetBySource(string(domain.SourceFeed), id)
	for _, a := range atts {
		feed.Attachments = append(feed.Attachments, *a)
	}

	return feed, nil
}

func (s *feedService) Update(id string, feed *domain.Feed, mediaIDs []string) error {
	feed.ID = id
	err := s.repo.Update(feed)
	if err != nil {
		return err
	}

	s.attService.UnlinkBySource(string(domain.SourceFeed), id)
	for _, mID := range mediaIDs {
		att := &domain.Attachment{
			SchoolID:   feed.SchoolID,
			SourceID:   id,
			SourceType: domain.SourceFeed,
			MediaID:    mID,
		}
		s.attService.Link(att)
	}
	return nil
}

func (s *feedService) Delete(id string) error {
	return s.repo.Delete(id)
}
