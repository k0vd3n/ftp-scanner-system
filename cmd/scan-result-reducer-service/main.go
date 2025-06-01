package main

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/logger"
	"ftp-scanner_try2/internal/mongodb"
	scanresultreducerservice "ftp-scanner_try2/internal/scan-result-reducer-service"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("Не удалось инициализировать логгер: " + err.Error())
	}

	zapLogger = zapLogger.With(zap.String("service", "scan-result-reducer-service"))
	zapLogger.Info("scan-result-reducer-service main: Запуск scan-result-reducer-service")
	zapLogger.Info("scan-result-reducer-service main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		zapLogger.Fatal("scan-result-reducer-service main: Ошибка загрузки конфига", zap.Error(err))
	}

	scanresultreducerservice.InitMetrics(cfg.ScanResultReducer.Metrics.InstanceLabel)

	go func() {
		zapLogger.Info("scan-result-reducer-service main: Запуск HTTP-сервера для метрик",
			zap.String("port", cfg.ScanResultReducer.Metrics.PromHttpPort))
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.ScanResultReducer.Metrics.PromHttpPort, nil); err != nil {
			zapLogger.Fatal("scan-result-reducer-service main: Ошибка HTTP-сервера для метрик", zap.Error(err))
		}
	}()

	zapLogger.Info("scan-result-reducer-service main: Подключение к MongoDB...",
		zap.String("uri", cfg.ScanResultReducer.Mongo.MongoUri),
		zap.String("database", cfg.ScanResultReducer.Mongo.MongoDb),
		zap.String("collection", cfg.ScanResultReducer.Mongo.MongoCollection))
	// Подключение к MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.ScanResultReducer.Mongo.MongoUri))
	if err != nil {
		zapLogger.Fatal("scan-result-reducer-service main: Ошибка подключения к MongoDB", zap.Error(err))
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			zapLogger.Error("scan-result-reducer-service main: Ошибка при отключении от MongoDB", zap.Error(err))
		}
	}()

	zapLogger.Info("scan-result-reducer-service main: Подключение к Kafka...")
	// Инициализация репозитория и сервиса
	repo := mongodb.NewMongoSaveReportRepository(
		client,
		cfg.ScanResultReducer.Mongo.MongoDb,
		cfg.ScanResultReducer.Mongo.MongoCollection,
		zapLogger,
	)

	zapLogger.Info("scan-result-reducer-service main: Инициализация Kafka-консьюмера...",
		zap.Any("brokers", cfg.ScanResultReducer.Kafka.Brokers),
		zap.String("topic", cfg.ScanResultReducer.Kafka.ConsumerTopic),
		zap.String("groupID", cfg.ScanResultReducer.Kafka.ConsumerGroup),
		zap.Int("batchSize", cfg.ScanResultReducer.Kafka.BatchSize),
		zap.Int("duration", cfg.ScanResultReducer.Kafka.Duration))

	consumer := kafka.NewScanResultConsumer(
		cfg.ScanResultReducer.Kafka.Brokers,
		cfg.ScanResultReducer.Kafka.ConsumerTopic,
		cfg.ScanResultReducer.Kafka.ConsumerGroup,
		zapLogger,
	)
	defer consumer.CloseReader()

	service := scanresultreducerservice.NewReducerService(repo, *cfg, consumer, zapLogger)
	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Горутинa для обработки сообщений
	zapLogger.Info("scan-result-reducer-service main: Запуск горутины для обработки сообщений")

	go service.Start(ctx)

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	cancel()
	zapLogger.Info("scan-result-reducer-service main: Завершение работы")
}
