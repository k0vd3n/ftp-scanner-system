package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ftp-scanner_try2/config"
	counterreducerservice "ftp-scanner_try2/internal/counter-reducer-service"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("counter-reducer-service main: Запуск сервиса редьюсирования...")
	log.Println("counter-reducer-service main: Загрузка конфига...")
	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("counter-reducer-service main: Ошибка загрузки конфига: %v", err)
	}

	// Инициализация MongoDB
	log.Println("counter-reducer-service main: Инициализация соединения с MongoDB...")
	mongoURI := cfg.CounterReducer.Mongo.MongoUri
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("counter-reducer-service main: Ошибка подключения к MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	repo := mongodb.NewCounterRepository(
		client,
		cfg.CounterReducer.Mongo.MongoDb,
		cfg.CounterReducer.Mongo.MongoCollection,
	)
	service := counterreducerservice.NewCounterReducerService(repo)

	// Инициализация Kafka-консьюмера
	log.Println("counter-reducer-service main: Инициализация Kafka-консьюмера...")
	consumer := kafka.NewCounterConsumer(
		cfg.CounterReducer.Kafka.Brokers,
		cfg.CounterReducer.Kafka.CounterReducerTopic,
		cfg.CounterReducer.Kafka.CounterReducerGroup,
	)
	defer consumer.CloseReader()

	// Обработка сообщений
	log.Println("counter-reducer-service main: Начало обработки сообщений...")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			messages, err := consumer.ReadMessages(
				ctx,
				cfg.CounterReducer.Kafka.BatchSize,
				time.Duration(cfg.CounterReducer.Kafka.Duration)*time.Second,
			)
			if err != nil {
				log.Printf("counter-reducer-service main: Ошибка чтения сообщений: %v", err)
				continue
			}
			log.Printf("counter-reducer-service main: Получено %d сообщений", len(messages))
			if len(messages) != 0 {
				log.Println("counter-reducer-service main: Редьюс сообщений...")
				reduced := service.ReduceMessages(messages)
				log.Println("counter-reducer-service main: Вставка редьюсированных данных в MongoDB...")
				if err := repo.InsertReducedCounters(ctx, reduced); err != nil {
					log.Printf("counter-reducer-service main: Ошибка вставки данных в MongoDB: %v", err)
				}
				log.Printf("counter-reducer-service main: Сообщения отправлены в Mongodb")
			}

		}
	}()

	// Завершение работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	cancel()
	log.Println("counter-reducer-service main: Сервис завершил работу")
}
