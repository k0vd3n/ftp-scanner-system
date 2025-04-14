package reportservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"log"
	"time"
)

type ReportServerInterface interface {
	proto.ReportServiceServer
}

type ReportServer struct {
	proto.UnimplementedReportServiceServer
	service ReportServiceInterface
}

func NewReportServer(service *ReportService) ReportServerInterface {
	return &ReportServer{service: service}
}

func (s *ReportServer) GenerateReport(ctx context.Context, req *proto.ReportRequest) (*proto.ReportResponse, error) {
	log.Printf("scan-reports-service server: getReport: Получение отчета для scan_id=%s", req.GetScanId())
	ReportRequestsTotal.Inc()
	start := time.Now()
	reportURL, err := s.service.GenerateReport(ctx, req.GetScanId())
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	ReportProcessingDuration.Observe(duration.Seconds())
	log.Printf("scan-reports-service server: getReport: Создан отчет для scan_id=%s", req.GetScanId())
	return &proto.ReportResponse{
		ScanId:    req.GetScanId(),
		ReportUrl: reportURL,
	}, nil
}
