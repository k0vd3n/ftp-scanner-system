package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ftp-scanner_try2/config"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service"
	"ftp-scanner_try2/internal/file-scanner-service/handler"
	"ftp-scanner_try2/internal/file-scanner-service/scanner"
	"ftp-scanner_try2/internal/file-scanner-service/service"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Инициализируем логгер
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("File-scanner-service: main: Не удалось инициализировать логгер: " + err.Error())
	}
	defer zapLogger.Sync()

	// Добавляем информацию о сервисе
	zapLogger = zapLogger.With(zap.String("service", "file-scanner-service"))
	zapLogger.Info("File-scanner-service: main: Запуск сервиса сканирования файлов")

	// Загрузка конфигурации
	zapLogger.Info("Загрузка конфига")
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		zapLogger.Fatal("File-scanner-service: main: Ошибка загрузки конфига", zap.Error(err))
	}

	// Инициализация метрик
	filescannerservice.InitMetrics(cfg.FileScanner.Metrics.InstanceLabel)
	go func() {
		zapLogger.Info("File-scanner-service: main: Запуск HTTP-сервера для метрик", 
			zap.String("port", cfg.FileScanner.Metrics.PromHttpPort))
		
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.FileScanner.Metrics.PromHttpPort, nil); err != nil {
			zapLogger.Fatal("File-scanner-service: main: Ошибка HTTP-сервера для метрик", zap.Error(err))
		}
	}()

	// Инициализация сканеров
	scannerTypes := cfg.FileScanner.KafkaScanResultProducer.ScannerTypes
	scannerMap := make(map[string]scanner.FileScanner)
	for _, st := range scannerTypes {
		switch st {
		case "zero_bytes":
			scannerMap[st] = scanner.NewZeroBytesScanner()
		}
	}

	// Создание Kafka consumer
	zapLogger.Info("file-scanner-service: main: Создание Kafka consumer")
	consumer := kafka.NewFileConsumer(
		cfg.FileScanner.KafkaConsumer.Brokers,
		cfg.FileScanner.KafkaConsumer.ConsumerTopic,
		cfg.FileScanner.KafkaConsumer.ConsumerGroup,
	)
	defer consumer.CloseReader()

	// Создание Kafka scan result producer
	zapLogger.Info("file-scanner-service: main: Создание Kafka scan result producer")
	scanResultProducer, err := kafka.NewScanResultProducer(
		cfg.FileScanner.KafkaScanResultProducer.Broker,
		cfg.FileScanner.KafkaScanResultProducer.Routing,
	)
	if err != nil {
		zapLogger.Fatal("file-scanner-service: main: Ошибка создания Kafka scan result producer", zap.Error(err))
	}
	defer scanResultProducer.CloseWriter()

	// Создание Kafka counter producer
	zapLogger.Info("file-scanner-service: main: Создание Kafka counter producer")
	counterProducer, err := kafka.NewProducer(cfg.FileScanner.KafkaCompletedFilesCountProducer.Broker, zapLogger)
	if err != nil {
		zapLogger.Fatal("file-scanner-service: main: Ошибка создания Kafka counter producer", zap.Error(err))
	}
	defer counterProducer.CloseWriter()

	// Инициализация сервиса
	zapLogger.Info("file-scanner-service: main: Создание сервиса")
	fileScannerService := service.NewFileScannerService(
		scanResultProducer,
		counterProducer,
		cfg.FileScanner,
		scannerMap,
		zapLogger,
	)

	// Создание обработчика Kafka
	kafkaHandler := handler.NewKafkaHandler(fileScannerService, consumer, zapLogger)

	// Настройка graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск обработки сообщений
	ctx, cancel := context.WithCancel(context.Background())
	zapLogger.Info("file-scanner-service: main: Запуск обработки сообщений")
	go kafkaHandler.Start(ctx)

	// Ожидание сигнала завершения
	<-stop
	cancel()
	zapLogger.Info("file-scanner-service: main: Сервис завершил работу")
}