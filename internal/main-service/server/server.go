package server

import (
	"context"
	mainservice "ftp-scanner_try2/internal/main-service"
	"ftp-scanner_try2/internal/models"
	"log"
)

type MainServer struct {
	scanService    mainservice.ScanService
	reportService  mainservice.ReportService
	counterService mainservice.CounterService
}

func NewMainServer(scanService mainservice.ScanService, reportService mainservice.ReportService, counterService mainservice.CounterService) *MainServer {
	return &MainServer{
		scanService:    scanService,
		reportService:  reportService,
		counterService: counterService,
	}
}

func (s *MainServer) StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error) {
	log.Println("Starting scan...")
	return s.scanService.StartScan(ctx, req)
}

func (s *MainServer) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
	log.Printf("Fetching scan status for scan_id=%s", scanID)
	return s.counterService.GetScanStatus(ctx, scanID)
}

func (s *MainServer) GetReport(ctx context.Context, scanID string) (*models.ReportResponse, error) {
	log.Printf("Fetching report for scan_id=%s", scanID)
	return s.reportService.GetReport(ctx, scanID)
}