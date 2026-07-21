package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"encoding/json"
)

type LogService interface {
	Record(log *domain.Log) error
	GetBySchool(schoolID string, page int, limit int) ([]*domain.Log, int64, error)
	GetByUser(userID string, page int, limit int) ([]*domain.Log, int64, error)
	GetByCorrelationID(correlationID string) ([]*domain.Log, error)
	Log(actor domain.ActorContext, action string, entityType string, entityID *string, severity string, metadata any) error
}

type logService struct {
	repo repository.LogRepository
}

func NewLogService(repo repository.LogRepository) LogService {
	return &logService{repo: repo}
}

func (s *logService) Record(log *domain.Log) error {
	return s.repo.Create(log)
}

func (s *logService) GetBySchool(schoolID string, page int, limit int) ([]*domain.Log, int64, error) {
	return s.repo.GetBySchool(schoolID, page, limit)
}

func (s *logService) GetByUser(userID string, page int, limit int) ([]*domain.Log, int64, error) {
	return s.repo.GetByUser(userID, page, limit)
}

func (s *logService) GetByCorrelationID(correlationID string) ([]*domain.Log, error) {
	return s.repo.GetByCorrelationID(correlationID)
}

func (s *logService) Log(actor domain.ActorContext, action string, entityType string, entityID *string, severity string, metadata any) error {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	metadataStr := string(metadataBytes)
	scope := actor.Scope

	entry := &domain.Log{
		SchoolID:          actor.SchoolID,
		UserID:            actor.UserID,
		ActorSchoolUserID: actor.SchoolUserID,
		Action:            action,
		Metadata:          metadataStr,
		EntityType:        &entityType,
		EntityID:          entityID,
		Scope:             &scope,
		Severity:          &severity,
	}

	return s.Record(entry)
}
