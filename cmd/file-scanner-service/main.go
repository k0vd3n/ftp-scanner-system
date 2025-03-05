package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/file-scanner-service/handler"
	"ftp-scanner_try2/internal/file-scanner-service/scanner"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service/service"
	"ftp-scanner_try2/internal/kafka"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения из .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Конфигурация Kafka consumer
	consumerBrokers := []string{os.Getenv("KAFKA_FILE_SCAN_SVC_BROKER")}
	topic := os.Getenv("KAFKA_FILE_SCAN_SVC_TOPIC")
	groupID := os.Getenv("KAFKA_FILE_SCAN_SVC_GROUP_ID")

	// Конфигурация Kafka producer
	producerBroker := os.Getenv("KAFKA_FILE_SCAN_SVC_BROKER")
	scanResultsTopic := os.Getenv("KAFKA_FILE_SCAN_RESULTS_TOPIC")
	completedFilesCountTopic := os.Getenv("KAFKA_COMPLETED_FILES_COUNT_TOPIC")

	// Папка для скачанных файлов и права доступа
	downloadPath := os.Getenv("FILE_SCAN_DOWNLOAD_PATH")
	permission := os.FileMode(0755)

	// Инициализация consumer и producer
	consumerConfig := config.KafkaConsumerConfig{
		Brokers: consumerBrokers,
		Topic:   topic,
		GroupId: groupID,
	}

	scannerTypesMap := map[string]scanner.FileScanner{
		"zero_bytes": scanner.NewZeroBytesScanner(),
	}

	producerConfig := config.FilesScannerConfig{
		Broker:                   producerBroker,
		ScanResultsTopic:         scanResultsTopic,
		CompletedFilesCountTopic: completedFilesCountTopic,
		PathForDownloadedFiles:   downloadPath,
		Permision:                permission,
		ScannerTypesMap:          scannerTypesMap,
	}

	consumer := kafka.NewFileConsumer(consumerConfig.Brokers, consumerConfig.Topic, consumerConfig.GroupId)
	defer consumer.CloseReader()

	producer, err := kafka.NewProducer(producerConfig.Broker, producerConfig.ScanResultsTopic)
	if err != nil {
		log.Fatalf("Ошибка создания Kafka producer: %v", err)
	}
	defer producer.CloseWriter()

	// Создаем сервис сканирования файлов
	fileScannerService := filescannerservice.NewFileScannerService(producer, producerConfig)

	// Создаем Kafka handler
	kafkaHandler := handler.NewKafkaHandler(fileScannerService, consumer)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем обработку сообщений
	ctx, cancel := context.WithCancel(context.Background())
	go kafkaHandler.Start(ctx)

	// Ожидание завершения
	<-stop
	cancel()
	log.Println("file-scanner-service завершил работу")
}
