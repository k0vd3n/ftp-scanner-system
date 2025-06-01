package main

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	generatereportservice "ftp-scanner_try2/internal/generate-report-service"
	"ftp-scanner_try2/internal/logger"
	"ftp-scanner_try2/internal/mongodb"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("Generate Report Service: main: Не удалось инициализировать логгер: " + err.Error())
	}
	defer zapLogger.Sync()

	zapLogger = zapLogger.With(zap.String("service", "generate-report-service"))
	zapLogger.Info("Generate Report Service: main: Запуск Generate Report Service")
	zapLogger.Info("Generate Report Service: main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		zapLogger.Fatal("Generate Report Service: main: Ошибка загрузки конфига", zap.Error(err))
	}

	generatereportservice.InitMetrics(cfg.GenerateReportService.Metrics.InstanceLabel)
	go func() {
		zapLogger.Info("Generate Report Service: main: Запуск HTTP-сервера для метрик", zap.String("port", cfg.GenerateReportService.Metrics.PromHttpPort))
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.GenerateReportService.Metrics.PromHttpPort, nil); err != nil {
			log.Fatalf("Ошибка HTTP-сервера для метрик: %v", err)
		}
	}()

	zapLogger.Info("Generate Report Service: main: Инициализация соединения с MongoDB...")
	zapLogger.Info("Generate Report Service: main: MongoDB URI: " + cfg.GenerateReportService.Mongo.MongoUri)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.GenerateReportService.Mongo.MongoUri))
	if err != nil {
		zapLogger.Fatal("Generate Report Service: main: Ошибка подключения к MongoDB", zap.Error(err))
	}
	defer client.Disconnect(context.TODO())

	zapLogger.Info("Generate Report Service: main: Инициализация соединения с Redis...")
	repo := mongodb.NewMongoReportRepository(
		client,
		cfg.GenerateReportService.Mongo.MongoDb,
		cfg.GenerateReportService.Mongo.MongoCollection1,
		cfg.GenerateReportService.Mongo.MongoCollection2,
		zapLogger,
	)
	storage := generatereportservice.NewFileReportStorage(cfg.GenerateReportService.Repository.Directory, zapLogger)

	service := generatereportservice.NewGenerateReportService(repo, storage, zapLogger)
	server := generatereportservice.NewGenerateReportServer(service, zapLogger)

	zapLogger.Info("Generate Report Service: main: Инициализация gRPC соединений...")
	grpcServer := grpc.NewServer()
	proto.RegisterReportServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", cfg.GenerateReportService.Grpc.ServerPort)
	if err != nil {
		zapLogger.Fatal("Generate Report Service: main: Ошибка создания TCP-сокета", zap.Error(err))
	}
	zapLogger.Info("Generate Report Service: main: Инициализация gRPC соединений...")

	go grpcServer.Serve(lis)

	// Настройка graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	grpcServer.GracefulStop()
}
