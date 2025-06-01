package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// Инициализируем логгер и добавляем служебное поле для сервиса
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("Не удалось инициализировать логгер: " + err.Error())
	}
	// Добавляем информацию о сервисе (аналог заголовка старых логов)
	zapLogger = zapLogger.With(zap.String("service", "directory-lister-service"))
	defer zapLogger.Sync()

	zapLogger.Info("Main-service: main: Запуск Main-service")
	zapLogger.Info("Main-service: main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}
	mainservice.InitMetrics(cfg.MainService.Metrics.InstanceLabel)
	// mainservice.StartPushLoop(&cfg.PushGateway)
	go func() {
		log.Printf("Запуск HTTP-сервера для метрик на порту %s", cfg.MainService.Metrics.PromHttpPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.MainService.Metrics.PromHttpPort, nil); err != nil {
			log.Fatalf("Ошибка HTTP-сервера для метрик: %v", err)
		}
	}()

	grpcGenerateReportServerAddress := cfg.MainService.GRPC.GeneateReportServerAddress + cfg.MainService.GRPC.GenerateReportServerPort
	grpcGetReportServerAddress := cfg.MainService.GRPC.GetReportServerAddress + cfg.MainService.GRPC.GetReportServerPort
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

	generateReportConn, err := grpc.NewClient(grpcGenerateReportServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка соединения с Report Service: %v", err)
	}
	defer generateReportConn.Close()

	getReportConn, err := grpc.NewClient(grpcGetReportServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Main-service: main: Ошибка соединения с Report Service: %v", err)
	}
	defer getReportConn.Close()

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
	scanService := service.NewKafkaScanService(kafkaProducer, configScanDirTopic, zapLogger)
	generateReportService := service.NewGRPCReportService(proto.NewReportServiceClient(generateReportConn), zapLogger)
	getReportService := service.NewGRPCGetReportService(proto.NewGetReportServiceClient(getReportConn), zapLogger)
	statusService := service.NewGRPCStatusService(proto.NewStatusServiceClient(statusConn), zapLogger)

	log.Printf("Main-service: main: Инициализация MainServer...")
	// Создаем MainServer
	server := mainservice.NewMainServer(scanService, generateReportService, getReportService, statusService, zapLogger, cfg.MainService)

	log.Printf("Main-service: main: Настройка роутера...")
	// Настройка роутера
	router := server.SetupRouter()

	log.Printf("Main-service: main: Запуск HTTP сервера...")
	// Запуск HTTP сервера
	log.Printf("Main-service: main: Сервер запущен на порту %s", cfg.MainService.HTTP.Port)
	if err := http.ListenAndServe(cfg.MainService.HTTP.Port, router); err != nil {
		log.Fatalf("Main-service: main: Ошибка запуска HTTP сервера: %v", err)
	}

	// Настройка graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
