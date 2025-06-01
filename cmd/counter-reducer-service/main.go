package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ftp-scanner_try2/config"
	counterreducerservice "ftp-scanner_try2/internal/counter-reducer-service"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/logger"
	"ftp-scanner_try2/internal/mongodb"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("counter-reducer-service main: Не удалось инициализировать логгер: " + err.Error())
	}
	defer zapLogger.Sync()

	zapLogger = zapLogger.With(zap.String("service", "counter-reducer-service"))
	zapLogger.Info("counter-reducer-service main: Запуск counter-reducer-service")
	zapLogger.Info("counter-reducer-service main: Загрузка конфигурации...")
	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		zapLogger.Fatal("counter-reducer-service main: Ошибка загрузки конфига", zap.Error(err))
	}

	counterreducerservice.InitMetrics(cfg.CounterReducer.Metrics.InstanceLabel)
	go func() {
		zapLogger.Info("counter-reducer-service main: Запуск HTTP-сервера для метрик",
			zap.String("port", cfg.CounterReducer.Metrics.PromHttpPort))
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.CounterReducer.Metrics.PromHttpPort, nil); err != nil {
			zapLogger.Fatal("counter-reducer-service main: Ошибка HTTP-сервера для метрик", zap.Error(err))
		}
	}()

	// Инициализация MongoDB
	zapLogger.Info("counter-reducer-service main: Инициализация соединения с MongoDB...")
	mongoURI := cfg.CounterReducer.Mongo.MongoUri
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		zapLogger.Fatal("counter-reducer-service main: Ошибка подключения к MongoDB", zap.Error(err))
	}
	defer client.Disconnect(context.TODO())

	repo := mongodb.NewCounterRepository(
		client,
		cfg.CounterReducer.Mongo.MongoDb,
		cfg.CounterReducer.Mongo.MongoCollection,
		zapLogger,
	)

	// Инициализация Kafka-консьюмера
	zapLogger.Info("counter-reducer-service main: Инициализация соединения с Kafka...")
	consumer := kafka.NewCounterConsumer(
		cfg.CounterReducer.Kafka.Brokers,
		cfg.CounterReducer.Kafka.CounterReducerTopic,
		cfg.CounterReducer.Kafka.CounterReducerGroup,
		zapLogger,
	)
	defer consumer.CloseReader()
	service := counterreducerservice.NewCounterReducerService(repo, consumer, *cfg)
	// Обработка сообщений
	zapLogger.Info("counter-reducer-service main: Инициализация Kafka-консьюмера...")
	ctx, cancel := context.WithCancel(context.Background())

	go service.Start(ctx)

	// Завершение работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	cancel()
	zapLogger.Info("counter-reducer-service main: Завершение работы...")
}
