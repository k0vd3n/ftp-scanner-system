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
	log.Printf("Fetching scan status for scan_id=%s", scanID)

	counters, err := s.client.GetCounters(ctx, &proto.CounterRequest{ScanId: scanID})
	if err != nil {
		log.Printf("Failed to fetch scan status for scan_id=%s: %v", scanID, err)
		return nil, err
	}

	directoriesScanned := int(counters.GetDirectoriesCount())
	filesScanned := int(counters.GetFilesCount())
	directoriesFound := int(counters.GetDirectoriesCount())
	filesFound := int(counters.GetFilesCount())

	if directoriesScanned == directoriesFound && filesScanned == filesFound {
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
