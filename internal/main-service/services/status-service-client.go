package service

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"
	"time"

	"go.uber.org/zap"
)

type GRPCCounterService struct {
	client proto.StatusServiceClient
	logger *zap.Logger
}

func NewGRPCStatusService(client proto.StatusServiceClient, logger *zap.Logger) *GRPCCounterService {
	return &GRPCCounterService{client: client, logger: logger}
}

func (s *GRPCCounterService) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
	s.logger.Info("Main-service: counter-client: getScanStatus: Попытка получения статуса сканирования для scan_id=" + scanID)

	counters, err := s.client.GetStatus(ctx, &proto.StatusRequest{ScanId: scanID})
	if err != nil {
		s.logger.Error("Main-service: counter-client: getScanStatus: Ошибка получения статуса сканирования для scan_id="+scanID, zap.Error(err))
		return nil, err
	}

	s.logger.Info("Main-service: counter-client: getScanStatus: Статус сканирования получен для scan_id="+scanID,
		zap.Any("counters", counters))

	directoriesFound := int(counters.GetDirectoriesCount())
	filesScanned := int(counters.GetCompletedFiles())
	directoriesScanned := int(counters.GetCompletedDirectories())
	filesFound := int(counters.GetFilesCount())
	s.logger.Info("Main-service: counter-client: getScanStatus: Сканирование завершено для scan_id="+scanID,
		zap.Any("directoriesFound", directoriesFound),
		zap.Any("filesScanned", filesScanned),
		zap.Any("directoriesScanned", directoriesScanned),
		zap.Any("filesFound", filesFound))

	if directoriesScanned == directoriesFound+1 && filesScanned == filesFound {
		s.logger.Info("Main-service: counter-client: getScanStatus: Сканирование завершено для scan_id=" + scanID)
		return &models.StatusResponse{
			ScanID:             scanID,
			Status:             "completed",
			DirectoriesScanned: directoriesScanned,
			FilesScanned:       filesScanned,
			DirectoriesFound:   directoriesFound + 1,
			FilesFound:         filesFound,
			StartTime:          time.Now().Format(time.RFC3339),
		}, nil
	} else {
		s.logger.Info("Main-service: counter-client: getScanStatus: Сканирование в процессе для scan_id=" + scanID)
		return &models.StatusResponse{
			ScanID:             scanID,
			Status:             "in_progress",
			DirectoriesScanned: directoriesScanned,
			FilesScanned:       filesScanned,
			DirectoriesFound:   directoriesFound + 1,
			FilesFound:         filesFound,
			StartTime:          time.Now().Format(time.RFC3339),
		}, nil
	}
}
