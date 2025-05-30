package main

import (
	"context"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/mongodb"
	scanresultreducerservice "ftp-scanner_try2/internal/scan-result-reducer-service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("scan-result-reducer-service main: Запуск сервиса редьюса результатов сканирования...")
	log.Println("scan-result-reducer-service main: Загрузка конфигурации...")

	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("scan-result-reducer-service main: Ошибка загрузки конфига: %v", err)
	}

	scanresultreducerservice.InitMetrics(cfg.ScanResultReducer.Metrics.InstanceLabel)

	go func() {
		log.Printf("Запуск HTTP-сервера для метрик на порту %s", cfg.ScanResultReducer.Metrics.PromHttpPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.ScanResultReducer.Metrics.PromHttpPort, nil); err != nil {
			log.Fatalf("Ошибка HTTP-сервера для метрик: %v", err)
		}
	}()

	log.Printf("scan-result-reducer-service main: Подключение к MongoDB")
	log.Printf("scan-result-reducer-service main: MongoDB URI: %s", cfg.ScanResultReducer.Mongo.MongoUri)
	log.Printf("scan-result-reducer-service main: MongoDB Database: %s", cfg.ScanResultReducer.Mongo.MongoDb)
	log.Printf("scan-result-reducer-service main: MongoDB Collection: %s", cfg.ScanResultReducer.Mongo.MongoCollection)
	// Подключение к MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.ScanResultReducer.Mongo.MongoUri))
	if err != nil {
		log.Fatalf("scan-result-reducer-service main: Ошибка подключения к MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("scan-result-reducer-service main: Ошибка при отключении от MongoDB: %v", err)
		}
	}()

	log.Printf("scan-result-reducer-service main: Инициалиция репозитория и сервиса...")
	// Инициализация репозитория и сервиса
	repo := mongodb.NewMongoSaveReportRepository(
		client,
		cfg.ScanResultReducer.Mongo.MongoDb,
		cfg.ScanResultReducer.Mongo.MongoCollection,
	)

	log.Printf("scan-result-reducer-service main: Инициализация Kafka-консьюмера...")
	// Инициализация Kafka-консьюмера
	log.Printf("scan-result-reducer-service main: Kafka brokers: %v", cfg.ScanResultReducer.Kafka.Brokers)
	log.Printf("scan-result-reducer-service main: Kafka topic: %s", cfg.ScanResultReducer.Kafka.ConsumerTopic)
	log.Printf("scan-result-reducer-service main: Kafka groupID: %s", cfg.ScanResultReducer.Kafka.ConsumerGroup)
	log.Printf("scan-result-reducer-service main: Kafka batchSize: %d", cfg.ScanResultReducer.Kafka.BatchSize)
	log.Printf("scan-result-reducer-service main: Kafka duration: %d", cfg.ScanResultReducer.Kafka.Duration)

	consumer := kafka.NewScanResultConsumer(
		cfg.ScanResultReducer.Kafka.Brokers,
		cfg.ScanResultReducer.Kafka.ConsumerTopic,
		cfg.ScanResultReducer.Kafka.ConsumerGroup,
	)
	defer consumer.CloseReader()

	service := scanresultreducerservice.NewReducerService(repo, *cfg, consumer)
	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Горутинa для обработки сообщений
	log.Println("scan-result-reducer-service main: Запуск обработки сообщений...")

	go service.Start(ctx)

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	cancel()
	log.Println("scan-result-reducer-service main: Сервис завершил работу")
}
