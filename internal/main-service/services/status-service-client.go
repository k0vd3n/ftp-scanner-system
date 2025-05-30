package service

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"
)

type GRPCCounterService struct {
	client proto.StatusServiceClient
}

func NewGRPCStatusService(client proto.StatusServiceClient) *GRPCCounterService {
	return &GRPCCounterService{client: client}
}

func (s *GRPCCounterService) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
	log.Printf("Main-service: counter-client: getScanStatus: Получение статуса сканирования для scan_id=%s", scanID)

	counters, err := s.client.GetStatus(ctx, &proto.StatusRequest{ScanId: scanID})
	if err != nil {
		log.Printf("MainService: counter-client: getScanStatus: Ошибка получения счетчиков для scan_id=%s: %v", scanID, err)
		return nil, err
	}

	log.Printf("Main-service: counter-client: getScanStatus: Счетчики получены для scan_id=%s: %+v", scanID, counters)

	directoriesFound := int(counters.GetDirectoriesCount())
	filesScanned := int(counters.GetCompletedFiles())
	directoriesScanned := int(counters.GetCompletedDirectories())
	filesFound := int(counters.GetFilesCount())
	log.Printf("Main-service: counter-client: getScanStatus: directoriesScanned=%d", directoriesScanned)
	log.Printf("Main-service: counter-client: getScanStatus: filesScanned=%d", filesScanned)
	log.Printf("Main-service: counter-client: getScanStatus: directoriesFound=%d", directoriesFound)
	log.Printf("Main-service: counter-client: getScanStatus: filesFound=%d", filesFound)

	if directoriesScanned == directoriesFound+1 && filesScanned == filesFound {
		log.Printf("Main-service: counter-client: getScanStatus: Сканирование завершено для scan_id=%s", scanID)
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
		log.Printf("Main-service: counter-client: getScanStatus: Сканирование в процессе для scan_id=%s", scanID)
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
