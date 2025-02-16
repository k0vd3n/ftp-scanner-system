package mainservice

import (
	"context"
	"ftp-scanner_try2/internal/kafka"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"
)

type KafkaScanService struct {
	kafkaProducer *kafka.Producer
}

func NewKafkaScanService(producer *kafka.Producer) *KafkaScanService {
	return &KafkaScanService{kafkaProducer: producer}
}

func (s *KafkaScanService) StartScan(ctx context.Context, req models.ScanRequest) (*models.ScanResponse, error) {
	scanID := generateScanID()

	msg := models.DirectoryScanMessage{
		ScanID:        scanID,
		DirectoryPath: req.DirectoryPath,
		ScanTypes:     req.ScanTypes,
	}

	// Логируем отправку сообщения в Kafka
	log.Printf("Sending scan request to Kafka: %v", msg)

	if err := s.kafkaProducer.SendMessage("directories-to-scan", msg); err != nil {
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
