package kafka

import (
	"context"
	"encoding/json"

	"ftp-scanner_try2/internal/models"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type DirectoryConsumer struct {
	Reader *kafka.Reader
	logger *zap.Logger
}

func NewDirectoryConsumer(brokers []string, topic, groupID string, logger *zap.Logger) KafkaDirectoryConsumerInterface {
	return &DirectoryConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		logger: logger,
	}
}

func (c *DirectoryConsumer) ReadMessage(ctx context.Context) (*models.DirectoryScanMessage, error) {
	c.logger.Info("Directory-consumer: Чтение сообщений из Kafka...")
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var scanMsg models.DirectoryScanMessage
	err = json.Unmarshal(msg.Value, &scanMsg)
	if err != nil {
		return nil, err
	}

	c.logger.Info("DirectoryConsumer: Получено сообщение",
		zap.Any("scanMsg", scanMsg))
	return &scanMsg, nil
}

func (c *DirectoryConsumer) CloseReader() error {
	c.logger.Info("DirectoryConsumer: Закрытие Kafka-консьюмера")

	// log.Println("Directory-consumer: Закрытие Kafka-консьюмера...")
	return c.Reader.Close()
}
