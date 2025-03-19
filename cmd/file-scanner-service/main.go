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
	filesToScanTopic := os.Getenv("KAFKA_FILE_SCAN_SVC_TOPIC")
	groupID := os.Getenv("KAFKA_FILE_SCAN_SVC_GROUP_ID")

	// Конфигурация Kafka producer
	producerBroker := os.Getenv("KAFKA_FILE_SCAN_SVC_BROKER")
	completedFilesCountTopic := os.Getenv("KAFKA_COMPLETED_FILES_COUNT_TOPIC")

	// Папка для скачанных файлов и права доступа
	downloadPath := os.Getenv("FILE_SCAN_DOWNLOAD_PATH")
	permission := os.FileMode(0755)

	// Инициализация consumer и producer
	consumerConfig := config.KafkaConsumerConfig{
		Brokers: consumerBrokers,
		Topic:   filesToScanTopic,
		GroupId: groupID,
	}

	// Создание мапы сканеров на основе ScannerTypes из конфига
	scannerTypes := []string{"zero_bytes"} // Пример, можно загрузить из переменных окружения
	scannerMap := make(map[string]scanner.FileScanner)
	for _, st := range scannerTypes {
		switch st {
		case "zero_bytes":
			scannerMap[st] = scanner.NewZeroBytesScanner()
		}
	}

	// allTopics := config.SizeBasedRouterTopics{
	// 	AllScanResultsTopic: os.Getenv("KAFKA_FILE_SCAN_RESULTS_TOPIC"),
	// 	EmptyResultTopic:    os.Getenv("KAFKA_TOPIC_EMPTY_RESULT"),
	// 	SmallResultTopic:    os.Getenv("KAFKA_TOPIC_SMALL_RESULT"),
	// 	MediumResultTopic:   os.Getenv("KAFKA_TOPIC_MEDIUM_RESULT"),
	// 	LargeResultTopic:    os.Getenv("KAFKA_TOPIC_LARGE_RESULT"),
	// }

	// Загрузка правил маршрутизации из конфига
	// routingConfig := config.RoutingConfig{
	// 	Rules: []config.RoutingRule{
	// 		{
	// 			ScanType:     "zero_bytes",
	// 			TriggerValue: "0",
	// 			OutputTopics: []string{os.Getenv("KAFKA_TOPIC_EMPTY_RESULT")},
	// 		},
	// 	},
	// 	DefaultTopic: os.Getenv("KAFKA_FILE_SCAN_RESULTS_TOPIC"),
	// }

	routingConfig, err := config.LoadRoutingConfig("config/routing.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига маршрутизации: %v", err)
	}

	producerConfig := config.FilesScannerConfig{
		Broker:                   producerBroker,
		CompletedFilesCountTopic: completedFilesCountTopic,
		PathForDownloadedFiles:   downloadPath,
		Permision:                permission,
		ScannerTypesMap:          scannerTypes,
		Routing:                  *routingConfig,
	}

	consumer := kafka.NewFileConsumer(consumerConfig.Brokers, consumerConfig.Topic, consumerConfig.GroupId)
	defer consumer.CloseReader()

	scanResultProducer, err := kafka.NewScanResultProducer(producerConfig.Broker, producerConfig.Routing)
	if err != nil {
		log.Fatalf("Ошибка создания Kafka scan result producer: %v", err)
	}
	defer scanResultProducer.CloseWriter()

	counterProducer, err := kafka.NewProducer(producerConfig.Broker)
	if err != nil {
		log.Fatalf("Ошибка создания Kafka counter producer: %v", err)
	}
	defer counterProducer.CloseWriter()

	// Создаем сервис
	fileScannerService := filescannerservice.NewFileScannerService(
		scanResultProducer,
		counterProducer,
		producerConfig,
		scannerMap,
	)

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
