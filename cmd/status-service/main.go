package main

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/logger"
	"ftp-scanner_try2/internal/mongodb"
	statusservice "ftp-scanner_try2/internal/status-service"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Инициализация логгера
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("Не удалось инициализировать логгер: " + err.Error())
	}

	zapLogger = zapLogger.With(zap.String("service", "status-service"))
	zapLogger.Info("Counter Service: main: Запуск Counter Service")
	zapLogger.Info("Counter Service: main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		zapLogger.Fatal("Counter Service: main: Ошибка загрузки конфига", zap.Error(err))
	}

	statusservice.InitMetrics(cfg.StatusService.Metrics.InstanceLabel)
	go func() {
		zapLogger.Info("Counter Service: main: Запуск HTTP-сервера для метрик", zap.String("port", cfg.StatusService.Metrics.PromHttpPort))
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.StatusService.Metrics.PromHttpPort, nil); err != nil {
			zapLogger.Fatal("Counter Service: main: Ошибка HTTP-сервера для метрик", zap.Error(err))
		}
	}()

	zapLogger.Info("Counter Service: main: Подключение к MongoDB...", zap.String("uri", cfg.StatusService.Mongo.MongoUri))

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.StatusService.Mongo.MongoUri))
	if err != nil {
		zapLogger.Fatal("Counter Service: main: Ошибка подключения к MongoDB", zap.Error(err))
	}
	defer client.Disconnect(context.TODO())

	repo := mongodb.NewMongoCounterRepository(client, cfg.StatusService.Mongo.MongoDb, zapLogger)
	service := statusservice.NewStatusService(repo, zapLogger)

	zapLogger.Info("Counter Service: main: Подключаемые коллекции",
		zap.String("scan_directories_count", cfg.StatusService.Mongo.ScanDirectoriesCount),
		zap.String("scan_files_count", cfg.StatusService.Mongo.ScanFilesCount),
		zap.String("completed_directories_count", cfg.StatusService.Mongo.CompletedDirectoriesCount),
		zap.String("completed_files_count", cfg.StatusService.Mongo.CompletedFilesCount))

	server := statusservice.NewStatusServer(service, cfg.StatusService.Mongo, zapLogger)

	lis, err := net.Listen("tcp", cfg.StatusService.Grpc.Port)
	if err != nil {
		zapLogger.Fatal("Counter Service: main: Ошибка прослушивания порта", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	proto.RegisterStatusServiceServer(grpcServer, server)

	zapLogger.Info("Counter Service: main: Сервис счетчиков запущен на порте", zap.String("port", cfg.StatusService.Grpc.Port))
	if err := grpcServer.Serve(lis); err != nil {
		zapLogger.Fatal("Counter Service: main: Ошибка запуска сервиса", zap.Error(err))
	}
}
