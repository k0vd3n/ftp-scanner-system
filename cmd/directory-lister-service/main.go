package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ftp-scanner_try2/config"
	directorylisterservice "ftp-scanner_try2/internal/directory-lister-service"
	"ftp-scanner_try2/internal/kafka"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.yaml", "Path to config file")
	flag.Parse()

	log.Printf("Directory Lister Service: main: Загрузка конфигурации...")
	cfg, err := config.LoadUnifiedConfig(configPath) // Исправлено здесь
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	// Инициализация метрик
	log.Printf("directory Lister Service: main: Инициализация метрик...")
	directorylisterservice.InitMetrics(cfg.DirectoryLister.Metrics.InstanceLabel)
	// Запускаем HTTP-сервер для экспорта метрик
	go func() {
		log.Printf("Запуск HTTP-сервера для метрик на порту %s", cfg.DirectoryLister.Metrics.PromHttpPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.DirectoryLister.Metrics.PromHttpPort, nil); err != nil {
			log.Fatalf("Ошибка HTTP-сервера для метрик: %v", err)
		}
	}()

	consumerBrokers := []string{cfg.DirectoryLister.Kafka.Broker}
	topic := cfg.DirectoryLister.Kafka.ConsumerTopic
	groupID := cfg.DirectoryLister.Kafka.ConsumerGroup

	producerConfig := cfg.DirectoryLister.Kafka

	log.Printf("directory Lister Service: main: создание консумера и продусера")
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
