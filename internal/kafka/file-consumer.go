package kafka

import (
	"context"
	"encoding/json"
	"ftp-scanner_try2/internal/models"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type FileConsumer struct {
	Reader *kafka.Reader
	logger *zap.Logger
}

func NewFileConsumer(brokers []string, topic, groupID string, logger *zap.Logger) KafkaFileConsumerInterface {
	return &FileConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		logger: logger,
	}
}

func (c *FileConsumer) ReadMessage(ctx context.Context) (*models.FileScanMessage, error) {
	c.logger.Info("File-consumer: Чтение сообщений из Kafka...")
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var scanMsg models.FileScanMessage
	err = json.Unmarshal(msg.Value, &scanMsg)
	if err != nil {
		c.logger.Error("File-consumer: Ошибка при разборе сообщения", zap.Error(err))
		return nil, err
	}

	c.logger.Info("File-consumer: Получено сообщение", zap.Any("scanMsg", scanMsg))
	return &scanMsg, nil
}

func (c *FileConsumer) CloseReader() error {
	c.logger.Info("File-consumer: Закрытие Kafka-консьюмера")
	return c.Reader.Close()
}
