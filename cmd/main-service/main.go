package main

import (
	"log"
	"net/http"

	"ftp-scanner_try2/api/grpc/proto"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/logger"
	mainservice "ftp-scanner_try2/internal/main-service"
	service "ftp-scanner_try2/internal/main-service/services"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Printf("Main-service: main: Запуск Main-service...")
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	// Инициализируем логгер и добавляем служебное поле для сервиса
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("Не удалось инициализировать логгер: " + err.Error())
	}
	// Добавляем информацию о сервисе (аналог заголовка старых логов)
	zapLogger = zapLogger.With(zap.String("service", "directory-lister-service"))
	defer zapLogger.Sync()

	mainservice.InitMetrics(cfg.MainService.Metrics.InstanceLabel)
	// mainservice.StartPushLoop(&cfg.PushGateway)
	go func() {
		log.Printf("Запуск HTTP-сервера для метрик на порту %s", cfg.MainService.Metrics.PromHttpPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.MainService.Metrics.PromHttpPort, nil); err != nil {
			log.Fatalf("Ошибка HTTP-сервера для метрик: %v", err)
		}
	}()

	grpcReportServerAddress := cfg.MainService.GRPC.ReportServerAddress + cfg.MainService.GRPC.ReportServerPort
	grpcStatusServerAddress := cfg.MainService.GRPC.StatusServerAddress + cfg.MainService.GRPC.StatusServerPort

	log.Printf("Main-service: main: Инициализация Kafka Producer...")
	// Инициализация Kafka Producer
	kafkaProducer, err := kafka.NewProducer(cfg.MainService.Kafka.Broker, zapLogger) // Адрес Kafka брокера
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

	statusConn, err := grpc.NewClient(grpcStatusServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка соединения с Status Service: %v", err)
	}
	defer statusConn.Close()

	log.Printf("Main-service: main: Инициализация сервисов...")
	// Создаем сервисы
	configScanDirTopic := config.DirectoryListerKafkaConfig{
		DirectoriesToScanTopic: cfg.MainService.Kafka.DirectoryTorpic,
	}
	scanService := service.NewKafkaScanService(kafkaProducer, configScanDirTopic)
	reportService := service.NewGRPCReportService(proto.NewReportServiceClient(reportConn))
	statusService := service.NewGRPCStatusService(proto.NewStatusServiceClient(statusConn))

	log.Printf("Main-service: main: Инициализация MainServer...")
	// Создаем MainServer
	server := mainservice.NewMainServer(scanService, reportService, statusService)

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
