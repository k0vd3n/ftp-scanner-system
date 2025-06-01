package statusservice

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"

	"go.uber.org/zap"
)

// Интерфейс сервиса счётчиков
type StatusService interface {
	GetCounters(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.StatusResponseGRPC, error)
}

// Реализация сервиса
type statusService struct {
	repo   mongodb.GetCounterRepository
	logger *zap.Logger
}

// Конструктор сервиса
func NewStatusService(repo mongodb.GetCounterRepository, logger *zap.Logger) StatusService {
	return &statusService{
		repo:   repo,
		logger: logger,
	}
}

// Метод получения счетчиков
func (s *statusService) GetCounters(ctx context.Context, scanID string, config config.StatusServiceMongo) (*models.StatusResponseGRPC, error) {
	s.logger.Info("MongoDB: Получение счетчиков по ID скана",
		zap.String("scanID", scanID),
		zap.String("scanFilesCount", config.ScanFilesCount),
		zap.String("scanDirectoriesCount", config.ScanDirectoriesCount),
		zap.String("completedDirectoriesCount", config.CompletedDirectoriesCount),
		zap.String("completedFilesCount", config.CompletedFilesCount))

	// Получение счетчиков
	counters, err := s.repo.GetCountersByScanID(ctx, scanID, config)
	if err != nil {
		s.logger.Error("MongoDB: Ошибка получения счетчиков по ID скана", zap.Error(err), zap.String("scanID", scanID))
		return nil, err
	}

	// Вывод счетчиков
	s.logger.Info("MongoDB: Счетчики получены для ID скана",
		zap.String("scanID", scanID),
		zap.Int64("files_count", counters.FilesCount),
		zap.Int64("directories_count", counters.DirectoriesCount),
		zap.Int64("completed_directories", counters.CompletedDirectories),
		zap.Int64("completed_files", counters.CompletedFiles))

	s.logger.Info("MongoDB: Счетчики получены для ID скана", zap.String("scanID", scanID), zap.Any("counters", counters))
	return counters, nil
}
