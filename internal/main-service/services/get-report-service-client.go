package service

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/internal/models"

	"go.uber.org/zap"
)

type GRPCGetReportService struct {
	client proto.GetReportServiceClient
	logger *zap.Logger
}

func NewGRPCGetReportService(client proto.GetReportServiceClient, logger *zap.Logger) *GRPCGetReportService {
	return &GRPCGetReportService{client: client, logger: logger}
}

func (s *GRPCGetReportService) GetReport(ctx context.Context, scanID string, dirPath string) (*models.DirectoryResponse, error) {
	s.logger.Info("Main-service: report-client: getReport: Попытка получения отчета для scan_id=" + scanID)

	grpcResp, err := s.client.GetDirectory(ctx, &proto.DirectoryRequest{ScanId: scanID, Path: dirPath})
	if err != nil {
		s.logger.Error("Main-service: report-client: getReport: Ошибка получения отчета для scan_id="+scanID, zap.Error(err))
		return &models.DirectoryResponse{}, err
	}

	result := models.DirectoryResponse{
		ScanID:    grpcResp.ScanId,
		Directory: protoToHTTPDir(grpcResp.GetDirectory()),
	}

	s.logger.Info("Main-service: report-client: getReport: Отчет получен для scan_id=" + scanID)

	return &result, nil
}

func protoToHTTPScanResult(in *proto.ScanResult) models.ScanResult {
	return models.ScanResult{Type: in.GetType(), Result: in.GetResult()}
}

func protoToHTTPFile(in *proto.File) models.File {
	out := models.File{Path: in.GetPath()}
	out.ScanResults = make([]models.ScanResult, len(in.GetScanResults()))
	for i, sr := range in.GetScanResults() {
		out.ScanResults[i] = protoToHTTPScanResult(sr)
	}
	return out
}

func protoToHTTPDir(in *proto.Directory) models.Directory {
	out := models.Directory{Directory: in.GetDirectory()}
	out.Subdirectory = make([]models.Directory, len(in.GetSubdirectory()))
	for i, sd := range in.GetSubdirectory() {
		out.Subdirectory[i] = protoToHTTPDir(sd)
	}
	out.Files = make([]models.File, len(in.GetFiles()))
	for i, f := range in.GetFiles() {
		out.Files[i] = protoToHTTPFile(f)
	}
	return out
}
