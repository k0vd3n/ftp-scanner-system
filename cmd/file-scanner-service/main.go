package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ftp-scanner_try2/config"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service"
	"ftp-scanner_try2/internal/file-scanner-service/handler"
	"ftp-scanner_try2/internal/file-scanner-service/scanner"
	"ftp-scanner_try2/internal/file-scanner-service/service"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/logger"
	"ftp-scanner_try2/internal/mongodb"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	instanceID := os.Getenv("POD_UID")
	if instanceID == "" {
		instanceID = os.Getenv("POD_NAME")
	}
	if instanceID == "" {
		instanceID = cfg.FileScanner.Metrics.InstanceLabel
	}
	if instanceID == "" {
		instanceID = "localhost"
	}
	zapLogger.Info("Using instanceID", zap.String("instance", instanceID))

	zapLogger.Info("File-scanner-service: main: Инициализация MongoDB")
	mongoOpts := options.Client().ApplyURI(cfg.FileScanner.Mongo.MongoUri)
	mongoClient, err := mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		zapLogger.Fatal("Не удалось подключиться к MongoDB", zap.Error(err))
	}
	defer mongoClient.Disconnect(context.Background())

	zapLogger.Info("File-scanner-service: main: Инициализация репозитория метрик")
	filescannerservice.InitMetrics(instanceID)
	metricsRepo := mongodb.NewMongoMetricRepository(
		mongoClient,
		cfg.FileScanner.Mongo.MongoDb,
		cfg.FileScanner.Mongo.MongoCollection,
		zapLogger,
	)

	// Пытаемся загрузить ранее сохранённые метрики
	zapLogger.Info("File-scanner-service: main: Пытаемся загрузить ранее сохранённые метрики")
	err = filescannerservice.LoadMetricsFromMongo(context.Background(), metricsRepo, instanceID)
	if err != nil {
		zapLogger.Warn("Не удалось загрузить метрики из MongoDB", zap.Error(err))
	} else {
		zapLogger.Info("Метрики успешно загружены из MongoDB")
	}

	go func() {
		zapLogger.Info("File-scanner-service: main: Запуск HTTP-сервера для метрик",
			zap.String("port", cfg.FileScanner.Metrics.PromHttpPort))
		handler := promhttp.Handler()
		http.Handle("/metrics", handler)
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
		case "filemeta":
			scannerMap[st] = scanner.NewFileMetaScanner()
		case "lines_counter":
			scannerMap[st] = scanner.NewChunkLineCounterScanner()
		}
	}

	// Создание Kafka consumer
	zapLogger.Info("file-scanner-service: main: Создание Kafka consumer")
	consumer := kafka.NewFileConsumer(
		cfg.FileScanner.KafkaConsumer.Brokers,
		cfg.FileScanner.KafkaConsumer.ConsumerTopic,
		cfg.FileScanner.KafkaConsumer.ConsumerGroup,
		zapLogger,
	)
	defer consumer.CloseReader()

	// Создание Kafka scan result producer
	zapLogger.Info("file-scanner-service: main: Создание Kafka scan result producer")
	scanResultProducer, err := kafka.NewScanResultProducer(
		cfg.FileScanner.KafkaScanResultProducer.Broker,
		cfg.FileScanner.KafkaScanResultProducer.Routing,
		zapLogger,
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

	// 2) Контекст, отменяемый при OS-сигналах
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	// Запуск обработки сообщений
	zapLogger.Info("file-scanner-service: main: Запуск обработки сообщений")

	go kafkaHandler.Start(internalCtx)
	select {
	case <-internalCtx.Done():
		zapLogger.Info("Внутренняя отмена, сохраняем метрики в MongoDB…")
		// внутренняя отмена: пушим метрики
		if err := filescannerservice.SaveMetricsToMongo(context.Background(), metricsRepo, instanceID); err != nil {
			zapLogger.Error("Ошибка сохранения метрик в MongoDB", zap.Error(err))
		} else {
			zapLogger.Info("Метрики успешно сохранены в MongoDB")
		}
		time.Sleep(15 * time.Second)
	case <-signalCtx.Done():
		// внешняя отмена (Ctrl+C или kill): ничего не пушим
		zapLogger.Info("Получен сигнал ОС, завершаем без отправки метрик")
	}

	// Ожидание сигнала завершения
	// <-stop
	// cancel()
	zapLogger.Info("file-scanner-service: main: Сервис завершил работу")
}
