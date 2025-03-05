package main

import (
	"context"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/mongodb"
	scanresultreducerservice "ftp-scanner_try2/internal/scan-result-reducer-service"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Загрузка переменных окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки .env: %v", err)
	}

	// Настройка Kafka
	brokers := []string{os.Getenv("KAFKA_BROKER")}
	topic := os.Getenv("KAFKA_FILE_SCAN_RESULTS_TOPIC")
	groupID := os.Getenv("KAFKA_REPORT_REDUCER_GROUP_ID")
	batchSize, err := strconv.Atoi(os.Getenv("KAFKA_BATCH_SIZE"))
	if err != nil {
		log.Fatalf("Ошибка преобразования переменной KAFKA_BATCH_SIZE: %v", err)
	}

	// Настройка MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DATABASE_REPORTS")
	mongoCollection := os.Getenv("MONGO_COLLECTION_REPORTS")

	// Подключение к MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Ошибка при отключении от MongoDB: %v", err)
		}
	}()

	// Инициализация репозитория и сервиса
	repo := mongodb.NewMongoSaveReportRepository(client, mongoDatabase, mongoCollection)
	service := scanresultreducerservice.NewReducerService()

	// Инициализация Kafka-консьюмера
	consumer := kafka.NewScanResultConsumer(brokers, topic, groupID)
	defer consumer.CloseReader()

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Горутинa для обработки сообщений
	go func() {
		for {
			messages, err := consumer.ReadMessages(ctx, batchSize)
			if err != nil {
				log.Printf("Ошибка чтения сообщений из Kafka: %v", err)
				continue
			}

			reducedResults := service.ReduceScanResults(messages)
			if err := repo.InsertScanReports(ctx, reducedResults); err != nil {
				log.Printf("Ошибка сохранения данных в MongoDB: %v", err)
			}
		}
	}()

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	cancel()
	log.Println("Сервис scan-result-reducer завершил работу")
}
