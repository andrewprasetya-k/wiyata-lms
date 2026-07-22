package service

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/events"
	"backend/internal/repository"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogService interface {
	Record(log *domain.Log) error
	GetBySchool(schoolID string, page int, limit int) ([]*domain.Log, int64, error)
	GetByUser(userID string, page int, limit int) ([]*domain.Log, int64, error)
	GetByCorrelationID(correlationID string) ([]*domain.Log, error)
	Log(actor domain.ActorContext, action string, entityType string, entityID *string, severity string, metadata any) error
	LogBatch(tx *gorm.DB, actor domain.ActorContext, parentAction string, parentEntityType string, parentEntityID *string, parentSeverity string, parentMetadata any, children []LogBatchChild) error
}

type LogBatchChild struct {
	Action     string
	EntityType string
	EntityID   *string
	Severity   string
	Metadata   any
}

type logService struct {
	repo        repository.LogRepository
	broadcaster events.AuditBroadcaster
}

func NewLogService(repo repository.LogRepository, broadcaster events.AuditBroadcaster) LogService {
	return &logService{repo: repo, broadcaster: broadcaster}
}

func (s *logService) Record(log *domain.Log) error {
	if err := s.repo.Create(log); err != nil {
		return err
	}
	s.publishAuditEvent(log)
	return nil
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
	entry, err := buildLogEntry(actor, action, entityType, entityID, severity, metadata, nil)
	if err != nil {
		return err
	}
	return s.Record(entry)
}

func (s *logService) LogBatch(tx *gorm.DB, actor domain.ActorContext, parentAction string, parentEntityType string, parentEntityID *string, parentSeverity string, parentMetadata any, children []LogBatchChild) error {
	correlationID := uuid.NewString()

	parent, err := buildLogEntry(actor, parentAction, parentEntityType, parentEntityID, parentSeverity, parentMetadata, &correlationID)
	if err != nil {
		return err
	}

	entries := make([]*domain.Log, 0, len(children)+1)
	entries = append(entries, parent)
	for _, child := range children {
		entry, err := buildLogEntry(actor, child.Action, child.EntityType, child.EntityID, child.Severity, child.Metadata, &correlationID)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
	}

	txRepo := s.repo.WithTx(tx)
	for _, entry := range entries {
		if err := txRepo.Create(entry); err != nil {
			return err
		}
	}
	s.publishAuditEvent(parent)
	return nil
}

type noopLogService struct{}

func (noopLogService) Record(*domain.Log) error { return nil }
func (noopLogService) GetBySchool(string, int, int) ([]*domain.Log, int64, error) {
	return nil, 0, nil
}
func (noopLogService) GetByUser(string, int, int) ([]*domain.Log, int64, error) {
	return nil, 0, nil
}
func (noopLogService) GetByCorrelationID(string) ([]*domain.Log, error) { return nil, nil }
func (noopLogService) Log(domain.ActorContext, string, string, *string, string, any) error {
	return nil
}
func (noopLogService) LogBatch(*gorm.DB, domain.ActorContext, string, string, *string, string, any, []LogBatchChild) error {
	return nil
}

func (s *logService) publishAuditEvent(log *domain.Log) {
	if s.broadcaster == nil || log == nil {
		return
	}

	item := dto.LogListItemDTO{
		ID:          log.ID,
		Action:      log.Action,
		ActorUserID: log.UserID,
		CreatedAt:   formatAPITime(log.CreatedAt),
	}
	if log.EntityType != nil {
		item.EntityType = *log.EntityType
	}
	if log.EntityID != nil {
		item.EntityID = *log.EntityID
	}
	if log.Scope != nil {
		item.Scope = *log.Scope
	}
	if log.Severity != nil {
		item.Severity = *log.Severity
	}
	if log.SchoolID != nil {
		item.SchoolID = *log.SchoolID
	}
	if log.CorrelationID != nil {
		item.CorrelationID = *log.CorrelationID
	}

	event := events.AuditEvent{Type: events.AuditEventTypeCreated, Payload: item}

	if log.Scope != nil && *log.Scope == domain.LogScopePlatform {
		event.Channel = "platform"
		s.broadcaster.BroadcastPlatformEvent(event)
		return
	}
	if log.SchoolID != nil {
		event.Channel = *log.SchoolID
		s.broadcaster.BroadcastSchoolEvent(*log.SchoolID, event)
	}
}

func buildLogEntry(actor domain.ActorContext, action string, entityType string, entityID *string, severity string, metadata any, correlationID *string) (*domain.Log, error) {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	metadataStr := string(metadataBytes)
	scope := actor.Scope

	return &domain.Log{
		SchoolID:          actor.SchoolID,
		UserID:            actor.UserID,
		ActorSchoolUserID: actor.SchoolUserID,
		Action:            action,
		Metadata:          metadataStr,
		EntityType:        &entityType,
		EntityID:          entityID,
		Scope:             &scope,
		Severity:          &severity,
		CorrelationID:     correlationID,
	}, nil
}
