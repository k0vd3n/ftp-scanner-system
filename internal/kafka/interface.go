package kafka

import (
	"context"
	"ftp-scanner_try2/internal/models"
)

type KafkaPoducerInterface interface {
	SendMessage(topic string, message interface{}) error
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

type KafkaCounterConsumerInterface interface {
	ReadMessages(ctx context.Context, batchSize int) ([]models.CountMessage, error)
	CloseReader() error
}

type KafkaScanResultConsumerInterface interface {
	ReadMessages(ctx context.Context, batchSize int) ([]models.ScanResultMessage, error)
	CloseReader() error
}