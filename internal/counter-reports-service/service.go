package counterservice

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
	"log"
)

// Интерфейс сервиса счётчиков
type CounterService interface {
	GetCounters(ctx context.Context, scanID string, config config.MongoCounterSvcConfig) (*models.CounterResponseGRPC, error)
}

// Реализация сервиса
type counterService struct {
	repo mongodb.GetCounterRepository
}

// Конструктор сервиса
func NewCounterService(repo mongodb.GetCounterRepository) CounterService {
	return &counterService{
		repo: repo,
	}
}

// Метод получения счетчиков
func (s *counterService) GetCounters(ctx context.Context, scanID string, config config.MongoCounterSvcConfig) (*models.CounterResponseGRPC, error) {
	// s.log.Infof("Fetching counters for scan_id: %s", scanID)
	log.Printf("Получаем счетчики для scan_id: %s", scanID)

	counters, err := s.repo.GetCountersByScanID(ctx, scanID, config)
	if err != nil {
		// s.log.Errorf("Failed to fetch counters for scan_id %s: %v", scanID, err)
		log.Printf("Ошибка получения счетчиков для scan_id %s: %v", scanID, err)
		return nil, err
	}

	// s.log.Infof("Counters fetched for scan_id %s: %+v", scanID, counters)
	log.Printf("Счетчики получены для scan_id %s: %+v", scanID, counters)
	return counters, nil
}
