package main

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/mongodb"
	statusservice "ftp-scanner_try2/internal/status-service"
	"log"
	"net"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	log.Printf("Counter Service: main: Запуск Counter Service...")
	log.Printf("Counter Service: main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Counter Service: main: Ошибка загрузки конфига: %v", err)
	}

	log.Printf("Counter Service: main: mongoUri: %s", cfg.StatusService.Mongo.MongoUri)

	log.Printf("Counter Service: main: mongoURI: %s", cfg.StatusService.Mongo.MongoUri)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.StatusService.Mongo.MongoUri))
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	repo := mongodb.NewMongoCounterRepository(client, cfg.StatusService.Mongo.MongoDb)
	service := statusservice.NewStatusService(repo)
	log.Printf("Counter Service: main: коллекция scan_directories_count: %s", cfg.StatusService.Mongo.ScanDirectoriesCount)
	log.Printf("Counter Service: main: коллекция scan_files_count: %s", cfg.StatusService.Mongo.ScanFilesCount)
	log.Printf("Counter Service: main: коллекция completed_directories_count: %s", cfg.StatusService.Mongo.CompletedDirectoriesCount)
	log.Printf("Counter Service: main: коллекция completed_files_count: %s", cfg.StatusService.Mongo.CompletedFilesCount)

	server := statusservice.NewStatusServer(service, cfg.StatusService.Mongo)

	lis, err := net.Listen("tcp", cfg.StatusService.Grpc.Port)
	if err != nil {
		log.Fatalf("Counter Service: main: Ошибка прослушивания порта: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterStatusServiceServer(grpcServer, server)

	log.Printf("Counter Service: main: Сервис счетчиков запущен на порте %s", cfg.StatusService.Grpc.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Counter Service: main: Ошибка запуска сервиса: %v", err)
	}
}
