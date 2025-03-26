package statusservice

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"log"
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
	log.Printf("counter-service server: GetCounters: Получаем счетчики для scan_id: %s", req.ScanId)

	counters, err := s.service.GetCounters(ctx, req.ScanId, s.config)
	if err != nil {
		log.Printf("counter-service server: GetCounters: Ошибка получения счетчиков для scan_id %s: %v", req.ScanId, err)
		return nil, err
	}

	log.Printf("counter-service server: GetCounters: Счетчик directories_count: %d", counters.DirectoriesCount)
	log.Printf("counter-service server: GetCounters: Счетчик files_count: %d", counters.FilesCount)
	log.Printf("counter-service server: GetCounters: Счетчик completed_directories: %d", counters.CompletedDirectories)
	log.Printf("counter-service server: GetCounters: Счетчик completed_files: %d", counters.CompletedFiles)
	return &proto.StatusResponse{
		ScanId:               counters.ScanID,
		DirectoriesCount:     counters.DirectoriesCount,
		FilesCount:           counters.FilesCount,
		CompletedDirectories: counters.CompletedDirectories,
		CompletedFiles:       counters.CompletedFiles,
	}, nil
}
