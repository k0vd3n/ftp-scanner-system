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

	return &models.StatusResponse{
		ScanID:             scanID,
		Status:             "in_progress",
		DirectoriesScanned: int(counters.GetDirectoriesCount()),
		FilesScanned:       int(counters.GetFilesCount()),
		DirectoriesFound:   int(counters.GetDirectoriesCount()),
		FilesFound:         int(counters.GetFilesCount()),
		StartTime:          "TODO: get from storage",
	}, nil
}