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
	GetCounters(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.CounterResponseGRPC, error)
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
func (s *counterService) GetCounters(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.CounterResponseGRPC, error) {
	log.Printf("counter-service service: Получаем счетчики для scan_id: %s", scanID)

	counters, err := s.repo.GetCountersByScanID(ctx, scanID, config)
	if err != nil {
		log.Printf("counter-service service: Ошибка получения счетчиков для scan_id %s: %v", scanID, err)
		return nil, err
	}

	log.Printf("counter-service service: Счетчик directories_count: %d", counters.DirectoriesCount)
	log.Printf("counter-service service: Счетчик files_count: %d", counters.FilesCount)
	log.Printf("counter-service service: Счетчик completed_directories: %d", counters.CompletedDirectories)
	log.Printf("counter-service service: Счетчик completed_files: %d", counters.CompletedFiles)

	// s.log.Infof("Counters fetched for scan_id %s: %+v", scanID, counters)
	log.Printf("counter-service service: Счетчики получены для scan_id %s: %+v", scanID, counters)
	return counters, nil
}
