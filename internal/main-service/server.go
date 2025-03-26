package mainservice

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"log"
)

type MainServer struct {
	scanServiceClient   ScanService
	reportServiceClient ReportService
	statusServiceClient StatusService
}

func NewMainServer(scanService ScanService, reportService ReportService, statusService StatusService) MainServiceInterface {
	return &MainServer{
		scanServiceClient:   scanService,
		reportServiceClient: reportService,
		statusServiceClient: statusService,
	}
}

// Остальные методы остаются без изменений
func (s *MainServer) StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error) {
	log.Println("Main-service: server: Начало сканирования...")

	response, err := s.scanServiceClient.StartScan(ctx, req)
	if err != nil {
		log.Printf("Main-service: server: Ошибка начала сканирования: %v", err)
		return nil, err
	}

	log.Printf("Main-service: server:  scan_id=%s", response.ScanID)
	return response, nil
}

func (s *MainServer) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
	log.Printf("Main-service: server: Получение статуса сканирования для scan_id=%s", scanID)

	response, err := s.statusServiceClient.GetScanStatus(ctx, scanID)
	if err != nil {
		log.Printf("Main-service: server: Ошибка получения статуса сканирования: %v", err)
		return nil, err
	}

	return response, nil
}

func (s *MainServer) GetReport(ctx context.Context, scanID string) (*models.ReportResponse, error) {
	log.Printf("Main-service: server: Получение отчета для scan_id=%s", scanID)

	response, err := s.reportServiceClient.GetReport(ctx, scanID)
	if err != nil {
		log.Printf("Main-service: server: Ошибка получения отчета: %v", err)
		return nil, err
	}

	return response, nil
}
