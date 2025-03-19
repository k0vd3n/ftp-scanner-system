package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	counterreducerservice "ftp-scanner_try2/internal/counter-reducer-service"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/mongodb"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env: %v", err)
	}

	// Настройка Kafka
	brokers := []string{os.Getenv("KAFKA_COUNTER_REDUCER_BROKER")}
	topic := os.Getenv("KAFKA_REDUCER_TOPIC")
	groupID := os.Getenv("KAFKA_COUNTER_GROUP_ID")
	batchSize, _ := strconv.Atoi(os.Getenv("COUNTER_REDUCER_BATCH_SIZE"))
	duration, _ := strconv.Atoi(os.Getenv("COUNTER_REDUCER_CYCLE_DURATION"))

	// Настройка MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	mongoDatabase := os.Getenv("MONGO_DATABASE_COUNTER")
	mongoCollection := os.Getenv("MONGO_COLLECTION_COUNTER")

	// Инициализация MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	repo := mongodb.NewCounterRepository(client, mongoDatabase, mongoCollection)
	service := counterreducerservice.NewCounterReducerService(repo)

	// Инициализация Kafka-консьюмера
	consumer := kafka.NewCounterConsumer(brokers, topic, groupID)
	defer consumer.CloseReader()

	// Обработка сообщений
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			messages, err := consumer.ReadMessages(ctx, batchSize, time.Duration(duration)*time.Second)
			if err != nil {
				log.Printf("Ошибка чтения сообщений: %v", err)
				continue
			}

			reduced := service.ReduceMessages(messages)
			if err := repo.InsertReducedCounters(ctx, reduced); err != nil {
				log.Printf("Ошибка вставки данных в MongoDB: %v", err)
			}
		}
	}()

	// Завершение работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	cancel()
	log.Println("Сервис завершил работу")
}
