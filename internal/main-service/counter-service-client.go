package mainservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"
	"log"
)

type GRPCCounterServiceClient struct {
	client proto.CounterServiceClient
}

func NewGRPCCounterService(client proto.CounterServiceClient) *GRPCCounterServiceClient {
	return &GRPCCounterServiceClient{client: client}
}

func (s *GRPCCounterServiceClient) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
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
