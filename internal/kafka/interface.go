package kafka

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"time"
)

type KafkaPoducerInterface interface {
	SendMessage(topic string, message interface{}) error
	CloseWriter() error
}

type KafkaScanResultPoducerInterface interface {
	SendMessage(message models.ScanResultMessage) error
	CloseWriter() error
}

type KafkaDirectoryConsumerInterface interface {
	ReadMessage(ctx context.Context) (*models.DirectoryScanMessage, error)
	CloseReader() error
}

type KafkaFileConsumerInterface interface {
	ReadMessage(ctx context.Context) (*models.FileScanMessage, error)
	CloseReader() error
}

type KafkaCounterReducerConsumerInterface interface {
	ReadMessages(ctx context.Context, batchSize int, duration time.Duration) ([]models.CountMessage, error)
	CloseReader() error
}

type KafkaScanResultReducerConsumerInterface interface {
	ReadMessages(ctx context.Context, batchSize int, duration time.Duration) ([]models.ScanResultMessage, error)
	CloseReader() error
}
