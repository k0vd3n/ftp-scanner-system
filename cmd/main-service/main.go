package main

import (
	"log"
	"net/http"

	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	mainservice "ftp-scanner_try2/internal/main-service"
	service "ftp-scanner_try2/internal/main-service/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Printf("Main-service: main: Запуск Main-service...")
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	grpcReportServerAddress := cfg.MainService.GRPC.ReportServerAddress + cfg.MainService.GRPC.ReportServerPort
	grpcCounterServerAddress := cfg.MainService.GRPC.CounterServerAddress + cfg.MainService.GRPC.CounterServerPort

	log.Printf("Main-service: main: Инициализация Kafka Producer...")
	// Инициализация Kafka Producer
	kafkaProducer, err := kafka.NewProducer(cfg.MainService.Kafka.Broker) // Адрес Kafka брокера
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка инициализации Kafka Producer: %v", err)
	}
	defer kafkaProducer.CloseWriter()

	log.Printf("Main-service: main: Инициализация gRPC соединений...")
	// Создаем gRPC соединения
	creds := insecure.NewCredentials()

	reportConn, err := grpc.NewClient(grpcReportServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка соединения с Report Service: %v", err)
	}
	defer reportConn.Close()

	counterConn, err := grpc.NewClient(grpcCounterServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка соединения с Counter Service: %v", err)
	}
	defer counterConn.Close()

	log.Printf("Main-service: main: Инициализация сервисов...")
	// Создаем сервисы
	configScanDirTopic := config.DirectoryListerKafkaConfig{
		DirectoriesToScanTopic: cfg.MainService.Kafka.DirectoryTorpic,
	}
	scanService := service.NewKafkaScanService(kafkaProducer, configScanDirTopic)
	reportService := service.NewGRPCReportService(proto.NewReportServiceClient(reportConn))
	counterService := service.NewGRPCCounterService(proto.NewCounterServiceClient(counterConn))

	log.Printf("Main-service: main: Инициализация MainServer...")
	// Создаем MainServer
	server := mainservice.NewMainServer(scanService, reportService, counterService)

	log.Printf("Main-service: main: Настройка роутера...")
	// Настройка роутера
	router := server.SetupRouter()

	log.Printf("Main-service: main: Запуск HTTP сервера...")
	// Запуск HTTP сервера
	log.Printf("Main-service: main: Сервер запущен на порту %s", cfg.MainService.HTTP.Port)
	if err := http.ListenAndServe(cfg.MainService.HTTP.Port, router); err != nil {
		log.Fatalf("Main-service: main: Ошибка запуска HTTP сервера: %v", err)
	}
}
