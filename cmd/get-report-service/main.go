package main

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	getreportservice "ftp-scanner_try2/internal/get-report-service"
	"ftp-scanner_try2/internal/logger"
	"ftp-scanner_try2/internal/mongodb"
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
		panic("Get Report Service: main: Не удалось инициализировать логгер: " + err.Error())
	}
	defer zapLogger.Sync()

	zapLogger = zapLogger.With(zap.String("service", "get-report-service"))
	zapLogger.Info("Get Report Service: main: Запуск Get Report Service")
	zapLogger.Info("Get Report Service: main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		zapLogger.Fatal("Get Report Service: main: Ошибка загрузки конфига", zap.Error(err))
	}

	getreportservice.InitMetrics(cfg.GetReportService.Metrics.InstanceLabel)
	go func() {
		zapLogger.Info("Get Report Service: main: Запуск HTTP-сервера для метрик",
			zap.String("port", cfg.GetReportService.Metrics.PromHttpPort))
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.GetReportService.Metrics.PromHttpPort, nil); err != nil {
			zapLogger.Fatal("Get Report Service: main: Ошибка HTTP-сервера для метрик", zap.Error(err))
		}
	}()

	zapLogger.Info("Get Report Service: main: Инициализация соединения с MongoDB...")
	zapLogger.Info("Get Report Service: main: MongoDB URI: " + cfg.GetReportService.Mongo.MongoUri)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.GetReportService.Mongo.MongoUri))
	if err != nil {
		zapLogger.Fatal("Get Report Service: main: Ошибка подключения к MongoDB", zap.Error(err))
	}
	defer client.Disconnect(context.TODO())

	zapLogger.Info("Get Report Service: main: Инициализация репозитория...")
	repo := mongodb.NewMongoGetReportRepository(
		client,
		cfg.GetReportService.Mongo.MongoDb,
		cfg.GetReportService.Mongo.MongoCollection,
		zapLogger,
	)

	service := getreportservice.NewGetReportService(repo, zapLogger)
	server := getreportservice.NewGetReportServer(service, zapLogger)

	zapLogger.Info("Get Report Service: main: Инициализация gRPC-сервера...")
	grpcServer := grpc.NewServer()
	proto.RegisterGetReportServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", cfg.GetReportService.Grpc.ServerPort)
	if err != nil {
		zapLogger.Fatal("Get Report Service: main: Ошибка создания TCP-сокета", zap.Error(err))
	}
	zapLogger.Info("Get Report Service: main: report-service запущен на порту", zap.String("port", cfg.GetReportService.Grpc.ServerPort))

	go grpcServer.Serve(lis)

	// Настройка graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	grpcServer.GracefulStop()
	zapLogger.Info("Get Report Service: main: Завершение работы")
}
