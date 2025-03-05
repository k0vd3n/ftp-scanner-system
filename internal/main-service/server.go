package mainservice

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"log"
)

/*
type MainServer struct {
	reportClient  proto.ReportServiceClient
	counterClient proto.CounterServiceClient
	kafkaProducer *kafka.Producer
}
*/

type MainServer struct {
	scanServiceClient    ScanService
	reportServiceClient  ReportService
	counterServiceClient CounterService
}

func NewMainServer(scanService ScanService, reportService ReportService, counterService CounterService) MainServiceInterface {
	return &MainServer{
		scanServiceClient:    scanService,
		reportServiceClient:  reportService,
		counterServiceClient: counterService,
	}
}

/*
// NewMainServer теперь принимает kafkaProducer как аргумент
func NewMainServer(kafkaProducer *kafka.Producer, grpcReportServerAddress string, grpcCounterServerAddress string) *MainServer {
	creds := insecure.NewCredentials()

	reportConn, err := grpc.NewClient(grpcReportServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to Report Service: %v", err)
	}

	counterConn, err := grpc.NewClient(grpcCounterServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to Counter Service: %v", err)
	}

	return &MainServer{
		reportClient:  proto.NewReportServiceClient(reportConn),
		counterClient: proto.NewCounterServiceClient(counterConn),
		kafkaProducer: kafkaProducer, // Теперь kafkaProducer определен
	}
}
*/
// Остальные методы остаются без изменений
func (s *MainServer) StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error) {
	log.Println("Starting scan...")

	response, err := s.scanServiceClient.StartScan(ctx, req)
	if err != nil {
		log.Printf("Failed to start scan: %v", err)
		return nil, err
	}

	log.Printf("Scan started: scan_id=%s", response.ScanID)
	return response, nil
}

func (s *MainServer) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
	log.Printf("Fetching scan status for scan_id=%s", scanID)

	response, err := s.counterServiceClient.GetScanStatus(ctx, scanID)
	if err != nil {
		log.Printf("Failed to get scan status: %v", err)
		return nil, err
	}

	return response, nil
}

func (s *MainServer) GetReport(ctx context.Context, scanID string) (*models.ReportResponse, error) {
	log.Printf("Fetching report for scan_id=%s", scanID)

	response, err := s.reportServiceClient.GetReport(ctx, scanID)
	if err != nil {
		log.Printf("Failed to get report: %v", err)
		return nil, err
	}

	return response, nil
}

/*
func (s *MainServer) GetCounters(ctx context.Context, scanID string) (*models.CounterResponseGRPC, error) {
	// Запрос счетчиков через GRPC
	counters, err := s.counterClient.GetCounters(ctx, &proto.CounterRequest{ScanId: scanID})
	if err != nil {
		return nil, err
	}

	return &models.CounterResponseGRPC{
		ScanID:               scanID,
		DirectoriesCount:     int(counters.GetDirectoriesCount()),
		FilesCount:           int(counters.GetFilesCount()),
		CompletedDirectories: int(counters.GetCompletedDirectories()),
		CompletedFiles:       int(counters.GetCompletedFiles()),
	}, nil
}
*/
