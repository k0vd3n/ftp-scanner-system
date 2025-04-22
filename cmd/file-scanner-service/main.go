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

	// Создание пушера для всех метрик
	pusher := filescannerservice.NewPusher(
		cfg.FileScanner.Metrics.PushGateway.URL,
		cfg.FileScanner.Metrics.PushGateway.JobName,
		cfg.FileScanner.Metrics.PushGateway.Instance,
	)

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


	// 1) Контекст для внутренней отмены (panic, ошибки и т.п.)
	internalCtx, internalCancel := context.WithCancel(context.Background())
	defer internalCancel()

	// Создание обработчика Kafka
	kafkaHandler := handler.NewKafkaHandler(fileScannerService, consumer, zapLogger, internalCancel)

	// Настройка graceful shutdown
	// stop := make(chan os.Signal, 1)
	// signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 2) Контекст, отменяемый при OS-сигналах
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	

	// Запуск обработки сообщений
	// ctx, cancel := context.WithCancel(context.Background())
	zapLogger.Info("file-scanner-service: main: Запуск обработки сообщений")

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		zapLogger.Error("file-scanner-service: main: PANIC RECOVERED произошло аварийное завершение", zap.Any("panic", r))
	// 	}
	// 	zapLogger.Info("file-scanner-service: main: Остановка обработки сообщений")
	// 	if err := pusher.Push(); err != nil {
	// 		zapLogger.Error("file-scanner-service: main: Ошибка отправки метрик", zap.Error(err))
	// 	}
	// }()

	go kafkaHandler.Start(internalCtx)
	select {
	case <-internalCtx.Done():
		// внутренняя отмена: пушим метрики
		if err := pusher.Push(); err != nil {
			zapLogger.Error("Ошибка отправки метрик при внутренней отмене", zap.Error(err))
		} else {
			zapLogger.Info("Метрики успешно отправлены после внутренней отмены")
		}
	case <-signalCtx.Done():
		// внешняя отмена (Ctrl+C или kill): ничего не пушим
		zapLogger.Info("Получен сигнал ОС, завершаем без отправки метрик")
	}
	
	// Ожидание сигнала завершения
	// <-stop
	// cancel()
	zapLogger.Info("file-scanner-service: main: Сервис завершил работу")
}
