package kafka

import (
	"context"
	"ftp-scanner_try2/internal/models"
)

type KafkaPoducerInterface interface {
	SendMessage(topic string, message interface{}) error
	CloseWriter() error
}

type KafkaConsumerInterface interface {
	ReadMessage(ctx context.Context) (*models.DirectoryScanMessage, error)
	CloseReader() error
}
