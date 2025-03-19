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

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ConsumerBrokers := []string{os.Getenv("KAFKA_DIR_LIST_SVC_BROKER")}
	topic := os.Getenv("KAFKA_DIR_LIST_SVC_TOPIC")
	groupID := os.Getenv("KAFKA_DIR_LIST_SVC_GROUP_ID")

	producerBroker := os.Getenv("KAFKA_DIR_LIST_SVC_BROKER")
	directoriesToScanTopic := os.Getenv("KAFKA_DIR_LIST_SVC_TOPIC")
	scanDirectoriesCountTopic := os.Getenv("KAFKA_DIR_LIST_SVC_TOPIC")
	filesToScanTopic := os.Getenv("KAFKA_DIR_LIST_SVC_TOPIC")
	scanFilesCountTopic := os.Getenv("KAFKA_DIR_LIST_SVC_TOPIC")
	completedDirectoriesCountTopic := os.Getenv("KAFKA_DIR_LIST_SVC_TOPIC")

	// ftpHost := os.Getenv("FTP_HOST")
	// ftpUser := os.Getenv("FTP_USER")
	// ftpPassword := os.Getenv("FTP_PASSWORD")

	consumerConfig := config.KafkaConsumerConfig{
		Brokers: ConsumerBrokers,
		Topic:   topic,
		GroupId: groupID,
	}

	producerConfig := config.DirectoryListerConfig{
		Broker:                         producerBroker,
		DirectoriesToScanTopic:         directoriesToScanTopic,
		ScanDirectoriesCountTopic:      scanDirectoriesCountTopic,
		FilesToScanTopic:               filesToScanTopic,
		ScanFilesCountTopic:            scanFilesCountTopic,
		CompletedDirectoriesCountTopic: completedDirectoriesCountTopic,
	}

	// ftpConfig := config.FTPConfig{
	// 	Host:     ftpHost,
	// 	User:     ftpUser,
	// 	Password: ftpPassword,
	// }

	// Создаем Kafka consumer и producer
	consumer := kafka.NewDirectoryConsumer(consumerConfig.Brokers, consumerConfig.Topic, consumerConfig.GroupId)
	defer consumer.CloseReader()

	producer, err := kafka.NewProducer(producerConfig.Broker, )
	if err != nil {
		log.Fatal("Ошибка создания Kafka producer:", err)
	}
	defer producer.CloseWriter()

	// Подключение к FTP серверу
	// ftpClient, err := ftpclient.NewFTPClient(ftpConfig.Host, ftpConfig.User, ftpConfig.Password)
	// if err != nil {
	// 	log.Fatal("Ошибка подключения к FTP:", err)
	// }
	// defer ftpClient.Close()
	// Создаем репозиторий и сервис

	service := directorylisterservice.NewDirectoryListerService(producer, producerConfig)
	// Создаем Kafka kafkaHandler
	kafkaHandler := directorylisterservice.NewKafkaHandler(service, consumer)

	// Канал для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск обработки сообщений
	ctx, cancel := context.WithCancel(context.Background())
	go kafkaHandler.Start(ctx, producerConfig)

	// Ожидание сигнала завершения
	<-stop
	cancel()
	log.Println("Сервис завершил работу")
}
