package reportservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"log"
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
	reportURL, err := s.service.GenerateReport(ctx, req.GetScanId())
	if err != nil {
		return nil, err
	}

	log.Printf("scan-reports-service server: getReport: Создан отчет для scan_id=%s", req.GetScanId())
	return &proto.ReportResponse{
		ScanId:    req.GetScanId(),
		ReportUrl: reportURL,
	}, nil
}
