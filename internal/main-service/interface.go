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
type GenerateReportService interface {
	GenerateReport(ctx context.Context, scanID string) (*models.ReportResponse, error)
}

type GetReportService interface {
	GetReport(ctx context.Context, scanID string, dirPath string) (*models.DirectoryResponse, error)
}

// Интерфейс сервиса счетчиков (gRPC Status Client)
type StatusService interface {
	GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error)
}

// Интерфейс для MainServer
type MainServiceInterface interface {
	StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error)
	GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error)
	GenerateReport(ctx context.Context, scanID string) (*models.ReportResponse, error)
	GetReport(ctx context.Context, scanID string, dirPath string) (*models.DirectoryResponse, error)

	HandleStartScan(w http.ResponseWriter, r *http.Request)
	HandleGetScanStatus(w http.ResponseWriter, r *http.Request)
	HandleGenerateReport(w http.ResponseWriter, r *http.Request)
	HandleGetReport(w http.ResponseWriter, r *http.Request)
	ApiHandleGetDirectory(w http.ResponseWriter, r *http.Request)

	SetupRouter() *http.ServeMux
}
