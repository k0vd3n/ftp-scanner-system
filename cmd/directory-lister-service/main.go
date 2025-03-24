package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ftp-scanner_try2/config"
	directorylisterservice "ftp-scanner_try2/internal/directory-lister-service"
	"ftp-scanner_try2/internal/kafka"
)

func main() {
	log.Printf("Directory Lister Service: main: Загрузка конфигурации...")
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	consumerBrokers := []string{cfg.DirectoryLister.Kafka.Broker}
	topic := cfg.DirectoryLister.Kafka.ConsumerTopic
	groupID := cfg.DirectoryLister.Kafka.ConsumerGroup

	producerConfig := cfg.DirectoryLister.Kafka

	// Создаем Kafka consumer и producer
	consumer := kafka.NewDirectoryConsumer(consumerBrokers, topic, groupID)
	defer consumer.CloseReader()

	producer, err := kafka.NewProducer(producerConfig.Broker)
	if err != nil {
		log.Fatal("Ошибка создания Kafka producer:", err)
	}
	defer producer.CloseWriter()

	service := directorylisterservice.NewDirectoryListerService(producer, producerConfig)
	// Создаем Kafka kafkaHandler
	kafkaHandler := directorylisterservice.NewKafkaHandler(service, consumer)

	// Канал для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск обработки сообщений
	ctx, cancel := context.WithCancel(context.Background())
	log.Printf("Directory Lister Service: main: Запуск обработки сообщений...")
	go kafkaHandler.Start(ctx, producerConfig)

	// Ожидание сигнала завершения
	<-stop
	cancel()
	log.Println("Сервис завершил работу")
}
