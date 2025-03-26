package service

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"
	"log"
)

type GRPCReportService struct {
	client proto.ReportServiceClient
}

func NewGRPCReportService(client proto.ReportServiceClient) *GRPCReportService {
	return &GRPCReportService{client: client}
}

func (s *GRPCReportService) GetReport(ctx context.Context, scanID string) (*models.ReportResponse, error) {
	log.Printf("Main-service: report-client: getReport: Получение отчета для scan_id=%s", scanID)
	
	report, err := s.client.GenerateReport(ctx, &proto.ReportRequest{ScanId: scanID})
	if err != nil {
		log.Printf("Main-service: report-client: getReport: Ошибка получения отчета для scan_id=%s: %v", scanID, err)
		return nil, err
	}

	log.Printf("Main-service: report-client: getReport: Отчет получен для scan_id=%s", scanID)

	return &models.ReportResponse{
		ScanID:    scanID,
		ReportURL: report.GetReportUrl(),
	}, nil
}