package main

import (
	"context"
	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/mongodb"
	reportservice "ftp-scanner_try2/internal/report-service"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	log.Printf("report-service: main: Запуск Report Service...")
	log.Printf("report-service: main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("counter-reducer-service main: Ошибка загрузки конфига: %v", err)
	}

	reportservice.InitMetrics(cfg.ReportService.Metrics.InstanceLabel)
	go func() {
		log.Printf("Запуск HTTP-сервера для метрик на порту %s", cfg.ReportService.Metrics.PromHttpPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.ReportService.Metrics.PromHttpPort, nil); err != nil {
			log.Fatalf("Ошибка HTTP-сервера для метрик: %v", err)
		}
	}()

	log.Printf("Report Service: main: Инициализация соединения с MongoDB...")
	log.Printf("Report Service: main: MongoDB URI: %s", cfg.ReportService.Mongo.MongoUri)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.ReportService.Mongo.MongoUri))
	if err != nil {
		log.Fatalf("Report Service: main: Ошибка подключения к MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	log.Printf("Report Service: main: Инициализация сервисов...")
	repo := mongodb.NewMongoReportRepository(
		client,
		cfg.ReportService.Mongo.MongoDb,
		cfg.ReportService.Mongo.MongoCollection,
	)
	storage := reportservice.NewFileReportStorage(cfg.ReportService.Repository.Directory)

	service := reportservice.NewReportService(repo, storage)
	server := reportservice.NewReportServer(service)

	log.Printf("Report Service: main: Инициализация gRPC соединений...")
	grpcServer := grpc.NewServer()
	proto.RegisterReportServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", cfg.ReportService.Grpc.ServerPort)
	if err != nil {
		log.Fatalf("Report Service: main: Ошибка создания TCP-сокета: %v", err)
	}
	log.Printf("Report Service: main: report-service запущен на порту %s", cfg.ReportService.Grpc.ServerPort)
	grpcServer.Serve(lis)
}
