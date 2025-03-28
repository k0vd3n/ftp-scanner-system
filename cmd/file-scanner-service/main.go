package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ftp-scanner_try2/config"
	filescannerservice "ftp-scanner_try2/internal/file-scanner-service"
	"ftp-scanner_try2/internal/file-scanner-service/handler"
	"ftp-scanner_try2/internal/file-scanner-service/scanner"
	"ftp-scanner_try2/internal/file-scanner-service/service"
	"ftp-scanner_try2/internal/kafka"
)

func main() {
	log.Println("file-scanner-service: Запуск сервиса сканирования файлов...")
	log.Println("file-scanner-service: Загрузка конфига...")
	// Загружаем unified config
	cfg, err := config.LoadUnifiedConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("file-scanner-service: Ошибка загрузки конфига: %v", err)
	}

	// Инициализация метрик для file-scanner-service
	filescannerservice.InitMetrics()
	filescannerservice.StartPushLoop(&cfg.PushGateway)

	scannerTypes := cfg.FileScanner.KafkaScanResultProducer.ScannerTypes
	scannerMap := make(map[string]scanner.FileScanner)
	for _, st := range scannerTypes {
		switch st {
		case "zero_bytes":
			scannerMap[st] = scanner.NewZeroBytesScanner()
		}
	}

	log.Printf("file-scanner-service: Создание Kafka consumer...")
	consumer := kafka.NewFileConsumer(
		cfg.FileScanner.KafkaConsumer.Brokers,
		cfg.FileScanner.KafkaConsumer.ConsumerTopic,
		cfg.FileScanner.KafkaConsumer.ConsumerGroup,
	)
	defer consumer.CloseReader()

	log.Printf("file-scanner-service: Создание Kafka scan result producer...")
	scanResultProducer, err := kafka.NewScanResultProducer(
		cfg.FileScanner.KafkaScanResultProducer.Broker,
		cfg.FileScanner.KafkaScanResultProducer.Routing,
	)
	if err != nil {
		log.Fatalf("file-scanner-service: Ошибка создания Kafka scan result producer: %v", err)
	}
	defer scanResultProducer.CloseWriter()

	log.Printf("file-scanner-service: Создание Kafka counter producer...")
	counterProducer, err := kafka.NewProducer(cfg.FileScanner.KafkaCompletedFilesCountProducer.Broker)
	if err != nil {
		log.Fatalf("file-scanner-service: Ошибка создания Kafka counter producer: %v", err)
	}
	defer counterProducer.CloseWriter()

	log.Printf("file-scanner-service: Создание сервиса...")
	// Создаем сервис
	fileScannerService := service.NewFileScannerService(
		scanResultProducer,
		counterProducer,
		cfg.FileScanner,
		scannerMap,
	)

	// Создаем Kafka handler
	kafkaHandler := handler.NewKafkaHandler(fileScannerService, consumer)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем обработку сообщений
	ctx, cancel := context.WithCancel(context.Background())
	log.Println("file-scanner-service: запуск обработки сообщений...")
	go kafkaHandler.Start(ctx)

	// Ожидание завершения
	<-stop
	cancel()
	log.Println("file-scanner-service: завершил работу")

}
