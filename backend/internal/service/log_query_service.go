package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
)

type LogQueryService interface {
	Search(filter repository.LogFilter) ([]*domain.Log, int64, error)
	GetByID(id string) (*domain.Log, error)
}

type logQueryService struct {
	repo repository.LogRepository
}

func NewLogQueryService(repo repository.LogRepository) LogQueryService {
	return &logQueryService{repo: repo}
}

func (s *logQueryService) Search(filter repository.LogFilter) ([]*domain.Log, int64, error) {
	return s.repo.Search(filter)
}

func (s *logQueryService) GetByID(id string) (*domain.Log, error) {
	return s.repo.GetByID(id)
}
