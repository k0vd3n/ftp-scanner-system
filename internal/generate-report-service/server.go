package reportservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"time"

	"go.uber.org/zap"
)

type GenerateReportServerInterface interface {
	proto.ReportServiceServer
}

type GenerateReportServer struct {
	proto.UnimplementedReportServiceServer
	service GenerateReportServiceInterface
	logger  *zap.Logger
}

func NewGenerateReportServer(service *GenerateReportService, logger *zap.Logger) GenerateReportServerInterface {
	return &GenerateReportServer{service: service, logger: logger}
}

func (s *GenerateReportServer) GenerateReport(ctx context.Context, req *proto.ReportRequest) (*proto.ReportResponse, error) {
	s.logger.Info("scan-reports-service server: getReport: Получение отчета для scan_id=" + req.GetScanId())
	ReportRequestsTotal.Inc()
	start := time.Now()
	err := s.service.GenerateReport(ctx, req.GetScanId())
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	ReportProcessingDuration.Observe(duration.Seconds())
	s.logger.Info("scan-reports-service server: getReport: Отчет получен для scan_id=" + req.GetScanId())
	return &proto.ReportResponse{
		ScanId: req.GetScanId(),
	}, nil
}
