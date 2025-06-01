package service

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"

	"go.uber.org/zap"
)

type GRPCGenerateReportService struct {
	client proto.ReportServiceClient
	logger *zap.Logger
}

func NewGRPCReportService(client proto.ReportServiceClient, logger *zap.Logger) *GRPCGenerateReportService {
	return &GRPCGenerateReportService{client: client, logger: logger}
}

func (s *GRPCGenerateReportService) GenerateReport(ctx context.Context, scanID string) (*models.ReportResponse, error) {
	s.logger.Info("Main-service: report-client: getReport: Попытка получения отчета для scan_id=" + scanID)

	_, err := s.client.GenerateReport(ctx, &proto.ReportRequest{ScanId: scanID})
	if err != nil {
		s.logger.Error("Main-service: report-client: getReport: Ошибка получения отчета для scan_id="+scanID, zap.Error(err))
		return nil, err
	}

	s.logger.Info("Main-service: report-client: getReport: Отчет получен для scan_id=" + scanID)

	return &models.ReportResponse{
		ScanID: scanID,
	}, nil
}
