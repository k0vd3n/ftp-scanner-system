package getreportservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"time"

	"go.uber.org/zap"
)

type GetReportServerInterface interface {
	GetDirectory(ctx context.Context, request *proto.DirectoryRequest) (*proto.DirectoryResponse, error)
}

type GetReportServer struct {
	proto.UnimplementedGetReportServiceServer
	service GetReportServiceInterface
	logger  *zap.Logger
}

func NewGetReportServer(service GetReportServiceInterface, logger *zap.Logger) *GetReportServer {
	return &GetReportServer{service: service, logger: logger}
}

func (s *GetReportServer) GetDirectory(ctx context.Context, request *proto.DirectoryRequest) (*proto.DirectoryResponse, error) {
	s.logger.Info("GetReportService: GetDirectory: Получение отчета", zap.String("scan_id", request.GetScanId()))
	ReportRequestsTotal.Inc()
	start := time.Now()
	response, err := s.service.GetDirectory(ctx, request)
	if err != nil {
		return nil, err
	}
	duration := time.Since(start)
	ReportProcessingDuration.Observe(duration.Seconds())
	s.logger.Info("GetReportService: GetDirectory: Отчет получен", zap.String("scan_id", request.GetScanId()))
	return response, nil
}
