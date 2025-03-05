package service

import (
	"context"
	"fmt"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"
)

type KafkaScanService struct {
	producer kafka.KafkaPoducerInterface
	config   config.DirectoryListerConfig
}

func NewKafkaScanService(producer kafka.KafkaPoducerInterface, config config.DirectoryListerConfig) *KafkaScanService {
	return &KafkaScanService{
		producer: producer,
		config:   config,
	}
}

func (s *KafkaScanService) StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error) {
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

	log.Printf("Sending scan request to Kafka: %v", msg)

	if err := s.producer.SendMessage(s.config.DirectoriesToScanTopic, msg); err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		return nil, err
	}

	log.Printf("Scan request sent successfully: scan_id=%s", scanID)

	return &models.ScanResponse{
		ScanID:  scanID,
		Status:  "accepted",
		Message: "Запрос на сканирование принят в обработку.",
		FTPConnection: models.FTPConnection{
			Server:   req.FTPServer,
			Port:     req.FTPPort,
			Username: req.FTPUsername,
		},
		StartTime: time.Now().Format(time.RFC3339),
	}, nil
}

func generateScanID() string {
	return fmt.Sprintf("scan-%d", time.Now().UnixNano())
}
