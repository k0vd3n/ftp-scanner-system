package statusservice

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
	"log"
)

// Интерфейс сервиса счётчиков
type StatusService interface {
	GetCounters(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.StatusResponseGRPC, error)
}

// Реализация сервиса
type statusService struct {
	repo mongodb.GetCounterRepository
}

// Конструктор сервиса
func NewStatusService(repo mongodb.GetCounterRepository) StatusService {
	return &statusService{
		repo: repo,
	}
}

// Метод получения счетчиков
func (s *statusService) GetCounters(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.StatusResponseGRPC, error) {
	log.Printf("counter-service service: Получаем счетчики для scan_id: %s из коллекций: %s, %s, %s, %s", 
	scanID, config.ScanFilesCount, config.ScanDirectoriesCount, config.CompletedDirectoriesCount, config.CompletedFilesCount)


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
