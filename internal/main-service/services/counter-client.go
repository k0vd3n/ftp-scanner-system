package service

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"
	"log"
)

type GRPCCounterService struct {
	client proto.CounterServiceClient
}

func NewGRPCCounterService(client proto.CounterServiceClient) *GRPCCounterService {
	return &GRPCCounterService{client: client}
}

func (s *GRPCCounterService) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
	log.Printf("Main-service: counter-client: getScanStatus: Получение статуса сканирования для scan_id=%s", scanID)

	counters, err := s.client.GetCounters(ctx, &proto.CounterRequest{ScanId: scanID})
	if err != nil {
		log.Printf("MainService: counter-client: getScanStatus: Ошибка получения счетчиков для scan_id=%s: %v", scanID, err)
		return nil, err
	}

	log.Printf("Main-service: counter-client: getScanStatus: Счетчики получены для scan_id=%s: %+v", scanID, counters)

	directoriesScanned := int(counters.GetDirectoriesCount())
	filesScanned := int(counters.GetFilesCount())
	directoriesFound := int(counters.GetCompletedDirectories())
	filesFound := int(counters.GetCompletedFiles())
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
			DirectoriesFound:   directoriesFound,
			FilesFound:         filesFound,
			StartTime:          "TODO: get from storage",
		}, nil
	} else {
		log.Printf("Main-service: counter-client: getScanStatus: Сканирование в процессе для scan_id=%s", scanID)
		return &models.StatusResponse{
			ScanID:             scanID,
			Status:             "in_progress",
			DirectoriesScanned: directoriesScanned,
			FilesScanned:       filesScanned,
			DirectoriesFound:   directoriesFound,
			FilesFound:         filesFound,
			StartTime:          "TODO: get from storage",
		}, nil
	}
}
