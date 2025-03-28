package statusservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"log"
	"time"
)

type StatusServerInterface interface {
	proto.StatusServiceServer
}

// структура для gRPC сервера
type StatusServer struct {
	proto.UnimplementedStatusServiceServer
	service StatusService
	config  config.StatusServiceMongo
}

// Конструктор сервера
func NewStatusServer(service StatusService, config config.StatusServiceMongo) StatusServerInterface {
	return &StatusServer{
		service: service,
		config:  config,
	}
}

// gRPC метод получения счетчиков
func (s *StatusServer) GetStatus(ctx context.Context, req *proto.StatusRequest) (*proto.StatusResponse, error) {
	log.Printf("status-service server: GetCounters: Получаем счетчики для scan_id: %s", req.ScanId)

	log.Printf("status-service server: GetCounters: Запрос получения счетчиков из коллекций %s, %s, %s, %s для scan_id=%s", 
	s.config.ScanFilesCount, s.config.ScanDirectoriesCount, s.config.CompletedDirectoriesCount, s.config.CompletedFilesCount, req.ScanId)
	RequestsTotal.Inc()
	overrallStart := time.Now()

	dbStart := time.Now()
	counters, err := s.service.GetCounters(ctx, req.ScanId, s.config)
	dbDuration := time.Since(dbStart).Seconds()
	DbQueryDuration.Observe(dbDuration)
	if err != nil {
		ErrorsTotal.Inc()
		log.Printf("status-service server: GetCounters: Ошибка получения счетчиков для scan_id %s: %v", req.ScanId, err)
		return nil, err
	}

	log.Printf("status-service server: GetCounters: Счетчик directories_count: %d", counters.DirectoriesCount)
	log.Printf("status-service server: GetCounters: Счетчик files_count: %d", counters.FilesCount)
	log.Printf("status-service server: GetCounters: Счетчик completed_directories: %d", counters.CompletedDirectories)
	log.Printf("status-service server: GetCounters: Счетчик completed_files: %d", counters.CompletedFiles)

	overallDuration := time.Since(overrallStart).Seconds()
	ProcessingDuration.Observe(overallDuration)
	log.Printf("status-service server: GetCounters: время обработки запроса %.4f сек", overallDuration)
	return &proto.StatusResponse{
		ScanId:               counters.ScanID,
		DirectoriesCount:     counters.DirectoriesCount,
		FilesCount:           counters.FilesCount,
		CompletedDirectories: counters.CompletedDirectories,
		CompletedFiles:       counters.CompletedFiles,
	}, nil
}
