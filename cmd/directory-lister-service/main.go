package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ftp-scanner_try2/config"
	directorylisterservice "ftp-scanner_try2/internal/directory-lister-service"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.yaml", "Path to config file")
	flag.Parse()

	// Инициализируем логгер и добавляем служебное поле для сервиса
	zapLogger, err := logger.InitLogger()
	if err != nil {
		panic("Не удалось инициализировать логгер: " + err.Error())
	}
	// Добавляем информацию о сервисе (аналог заголовка старых логов)
	zapLogger = zapLogger.With(zap.String("service", "directory-lister-service"))
	defer zapLogger.Sync()

	zapLogger.Info("Directory-lister-service: Загрузка конфигурации")
	cfg, err := config.LoadUnifiedConfig(configPath) // Исправлено здесь
	if err != nil {
		zapLogger.Fatal("Directory-lister-service: Ошибка загрузки конфигурации", zap.Error(err))
	}

	// Инициализация метрик
	zapLogger.Info("Directory-lister-service: Инициализация метрик")
	directorylisterservice.InitMetrics(cfg.DirectoryLister.Metrics.InstanceLabel)
	// Запускаем HTTP-сервер для экспорта метрик
	go func() {
		zapLogger.Info("Directory-lister-service: Запуск HTTP-сервера для метрик", zap.String("port", cfg.DirectoryLister.Metrics.PromHttpPort))
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.DirectoryLister.Metrics.PromHttpPort, nil); err != nil {
			zapLogger.Fatal("Directory-lister-service: Ошибка HTTP-сервера для метрик", zap.Error(err))
		}
	}()

	// Инициализация Jaeger трейсера
	// _, closer := tracer.InitJaeger(cfg.DirectoryLister.Jaeger)
	// defer closer.Close()

	consumerBrokers := []string{cfg.DirectoryLister.Kafka.Broker}
	topic := cfg.DirectoryLister.Kafka.ConsumerTopic
	groupID := cfg.DirectoryLister.Kafka.ConsumerGroup

	producerConfig := cfg.DirectoryLister.Kafka

	zapLogger.Info("Directory-lister-service: Создание Kafka consumer и producer")
	// Создаем Kafka consumer и producer
	consumer := kafka.NewDirectoryConsumer(consumerBrokers, topic, groupID, zapLogger)
	defer consumer.CloseReader()

	producer, err := kafka.NewProducer(producerConfig.Broker, zapLogger)
	if err != nil {
		zapLogger.Fatal("Directory-lister-service: Ошибка создания Kafka producer", zap.Error(err))
	}
	defer producer.CloseWriter()

	service := directorylisterservice.NewDirectoryListerService(producer, producerConfig, zapLogger)
	// Создаем Kafka kafkaHandler
	kafkaHandler := directorylisterservice.NewKafkaHandler(service, consumer, zapLogger)

	// Канал для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск обработки сообщений
	ctx, cancel := context.WithCancel(context.Background())
	zapLogger.Info("Directory-lister-service: Запуск обработки сообщений из Kafka")
	go kafkaHandler.Start(ctx, producerConfig)

	// Ожидание сигнала завершения
	<-stop
	cancel()
	zapLogger.Info("Directory-lister-service: Сервис завершил работу")
}
