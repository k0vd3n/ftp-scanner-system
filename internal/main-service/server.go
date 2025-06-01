package mainservice

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"log"

	"go.uber.org/zap"
)

type MainServer struct {
	scanServiceClient           ScanService
	generateReportServiceClient GenerateReportService
	getReportServiceClient      GetReportService
	statusServiceClient         StatusService
	logger                      *zap.Logger
	config                      config.MainServiceConfig
}

func NewMainServer(
	scanService ScanService,
	generateReportService GenerateReportService,
	getReportService GetReportService,
	statusService StatusService,
	logger *zap.Logger,
	config config.MainServiceConfig,
) MainServiceInterface {
	return &MainServer{
		scanServiceClient:           scanService,
		generateReportServiceClient: generateReportService,
		getReportServiceClient:      getReportService,
		statusServiceClient:         statusService,
		logger:                      logger,
		config:                      config,
	}
}

// Остальные методы остаются без изменений
func (s *MainServer) StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error) {
	s.logger.Info("Main-service: server: startScan: Начало сканирования", zap.Any("request", req))

	response, err := s.scanServiceClient.StartScan(ctx, req)
	if err != nil {
		s.logger.Error("Main-service: server: startScan: Ошибка начала сканирования", zap.Error(err))
		return nil, err
	}

	log.Printf("Main-service: server:  scan_id=%s", response.ScanID)
	return response, nil
}

func (s *MainServer) GetScanStatus(ctx context.Context, scanID string) (*models.StatusResponse, error) {
	s.logger.Info("Main-service: server: getScanStatus: Попытка получения статуса сканирования", zap.String("scan_id", scanID))

	response, err := s.statusServiceClient.GetScanStatus(ctx, scanID)
	if err != nil {
		s.logger.Error("Main-service: server: getScanStatus: Ошибка получения статуса сканирования", zap.Error(err))
		return nil, err
	}

	return response, nil
}

func (s *MainServer) GenerateReport(ctx context.Context, scanID string) (*models.ReportResponse, error) {
	s.logger.Info("Main-service: server: Генерация отчета для scan_id=" + scanID)

	response, err := s.generateReportServiceClient.GenerateReport(ctx, scanID)
	if err != nil {
		s.logger.Error("Main-service: server: Ошибка генерации отчета", zap.Error(err))
		return nil, err
	}

	return response, nil
}

func (s *MainServer) GetReport(ctx context.Context, scanID string, dirPath string) (*models.DirectoryResponse, error) {
	s.logger.Info("Main-service: server: getReport: Попытка получения отчета для scan_id=" + scanID)

	response, err := s.getReportServiceClient.GetReport(ctx, scanID, dirPath)
	if err != nil {
		s.logger.Error("Main-service: server: getReport: Ошибка получения отчета для scan_id="+scanID, zap.Error(err))
		return nil, err
	}

	return response, nil
}
