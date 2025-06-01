package service

import (
	"context"
	"fmt"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	mainservice "ftp-scanner_try2/internal/main-service"
	"ftp-scanner_try2/internal/models"
	"time"

	"go.uber.org/zap"
)

type KafkaScanService struct {
	producer kafka.KafkaPoducerInterface
	config   config.DirectoryListerKafkaConfig
	logger   *zap.Logger
}

func NewKafkaScanService(producer kafka.KafkaPoducerInterface, config config.DirectoryListerKafkaConfig, logger *zap.Logger) *KafkaScanService {
	return &KafkaScanService{
		producer: producer,
		config:   config,
		logger:   logger,
	}
}

func (s *KafkaScanService) StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error) {
	s.logger.Info("Main-service: scan: startScan: Начало сканирования", zap.Any("request", req))
	scanID := generateScanID()

	msg := models.DirectoryScanMessage{
		ScanID:        scanID,
		DirectoryPath: req.DirectoryPath,
		ScanTypes:     req.ScanTypes,
		FTPConnection: models.FTPConnection{
			Server:   req.FTPServer,
			Port:     req.FTPPort,
			Username: req.FTPUsername,
			Password: req.FTPPassword,
		},
	}

	s.logger.Info("Main-service: scan: startScan: Отправка Kafka сообщения", zap.Any("message", msg))

	startKafka := time.Now()
	if err := s.producer.SendMessage(s.config.DirectoriesToScanTopic, msg); err != nil {
		s.logger.Error("Main-service: scan: startScan: Ошибка отправки Kafka сообщения", zap.Error(err))
		return nil, err
	}
	durationKafka := time.Since(startKafka).Seconds()
	mainservice.KafkaPublishDuration.Observe(durationKafka)

	s.logger.Info("Main-service: scan: startScan: Сообщение отправлено в Kafka", zap.Any("message", msg))

	return &models.ScanResponse{
		ScanID:    scanID,
		Status:    "accepted",
		Message:   "Запрос на сканирование принят в обработку.",
		StartTime: time.Now().Format(time.RFC3339),
	}, nil
}

func generateScanID() string {
	return fmt.Sprintf("scan-%d", time.Now().UnixNano())
}
