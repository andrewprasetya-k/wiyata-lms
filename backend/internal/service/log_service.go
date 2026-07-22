package service

import (
	"backend/internal/domain"
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
	// LogBatch writes one parent row plus its child rows, all sharing a
	// single generated correlation_id (Phase 10.2 §5, Option B), inside the
	// caller-supplied transaction. For bulk actions (e.g. CSV import) that
	// must log in the same transaction as the business writes they describe.
	LogBatch(tx *gorm.DB, actor domain.ActorContext, parentAction string, parentEntityType string, parentEntityID *string, parentSeverity string, parentMetadata any, children []LogBatchChild) error
}

// LogBatchChild describes one child row of a LogBatch call. It shares the
// parent's actor and correlation_id; only its own action/entity/severity/
// metadata differ.
type LogBatchChild struct {
	Action     string
	EntityType string
	EntityID   *string
	Severity   string
	Metadata   any
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
	return nil
}

// noopLogService lets callers that don't care about audit logging (mainly
// unit tests constructing a business service directly) pass nil for
// LogService without every audit call site needing its own nil-check —
// mirrors the existing noopEmailService convention.
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
