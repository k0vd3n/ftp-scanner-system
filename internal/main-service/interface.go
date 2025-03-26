package mainservice

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"net/http"
)

// Интерфейс сервиса сканирования (Kafka Producer)
type ScanService interface {
	StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error)
}

// Интерфейс сервиса отчетов (gRPC Report Client)
type ReportService interface {
	GetReport(ctx context.Context, scanID string) (*models.ReportResponse, error)
}

// Интерфейс сервиса счетчиков (gRPC Status Client)
type StatusService interface {
	GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error)
}

// Интерфейс для MainServer
type MainServiceInterface interface {
	StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error)
	GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error)
	GetReport(ctx context.Context, scanID string) (*models.ReportResponse, error)
	SetupRouter() *http.ServeMux
}
