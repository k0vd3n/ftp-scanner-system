package reportservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
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
	reportURL, err := s.service.GenerateReport(ctx, req.GetScanId())
	if err != nil {
		return nil, err
	}

	return &proto.ReportResponse{
		ScanId:    req.GetScanId(),
		ReportUrl: reportURL,
	}, nil
}
