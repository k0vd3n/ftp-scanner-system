package statusservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"time"

	"go.uber.org/zap"
)

type StatusServerInterface interface {
	proto.StatusServiceServer
}

// структура для gRPC сервера
type StatusServer struct {
	proto.UnimplementedStatusServiceServer
	service StatusService
	config  config.StatusServiceMongo
	logger  *zap.Logger
}

// Конструктор сервера
func NewStatusServer(service StatusService, config config.StatusServiceMongo, logger *zap.Logger) StatusServerInterface {
	return &StatusServer{
		service: service,
		config:  config,
		logger:  logger,
	}
}

// gRPC метод получения счетчиков
func (s *StatusServer) GetStatus(ctx context.Context, req *proto.StatusRequest) (*proto.StatusResponse, error) {
	s.logger.Info("gRPC: Получение счетчиков", zap.String("scan_id", req.ScanId))

	s.logger.Info("gRPC: Получение счетчиков",
		zap.String("scan_id", req.ScanId),
		zap.String("scanFilesCount", s.config.ScanFilesCount),
		zap.String("scanDirectoriesCount", s.config.ScanDirectoriesCount),
		zap.String("completedDirectoriesCount", s.config.CompletedDirectoriesCount),
		zap.String("completedFilesCount", s.config.CompletedFilesCount))
	RequestsTotal.Inc()
	overrallStart := time.Now()

	dbStart := time.Now()
	counters, err := s.service.GetCounters(ctx, req.ScanId, s.config)
	dbDuration := time.Since(dbStart).Seconds()
	DbQueryDuration.Observe(dbDuration)
	if err != nil {
		ErrorsTotal.Inc()
		s.logger.Error("gRPC: Ошибка получения счетчиков", zap.Error(err), zap.String("scan_id", req.ScanId))
		return nil, err
	}

	s.logger.Info("gRPC: Счетчики получены для ID скана",
		zap.String("scan_id", req.ScanId),
		zap.Int64("directories_count", counters.DirectoriesCount),
		zap.Int64("files_count", counters.FilesCount),
		zap.Int64("completed_directories", counters.CompletedDirectories),
		zap.Int64("completed_files", counters.CompletedFiles))

	overallDuration := time.Since(overrallStart).Seconds()
	s.logger.Info("Измеренное время обработки запроса", zap.Float64("seconds", overallDuration))
	ProcessingDuration.Observe(overallDuration)
	s.logger.Info("status-service server: GetCounters: время обработки запроса", zap.Float64("seconds", overallDuration))
	return &proto.StatusResponse{
		ScanId:               counters.ScanID,
		DirectoriesCount:     counters.DirectoriesCount,
		FilesCount:           counters.FilesCount,
		CompletedDirectories: counters.CompletedDirectories,
		CompletedFiles:       counters.CompletedFiles,
	}, nil
}
