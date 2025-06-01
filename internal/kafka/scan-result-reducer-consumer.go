package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"ftp-scanner_try2/internal/models"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type ScanResultConsumer struct {
	Reader *kafka.Reader
	logger *zap.Logger
}

func NewScanResultConsumer(brokers []string, topic, groupID string, logger *zap.Logger) KafkaScanResultReducerConsumerInterface {
	return &ScanResultConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			CommitInterval: 0,
		}),
		logger: logger,
	}
}

func (c *ScanResultConsumer) ReadMessages(ctx context.Context, batchSize int, duration time.Duration) ([]models.ScanResultMessage, int, error) {
	c.logger.Info("scan-result-reducer-consumer: чтение сообщений из Kafka из топика",
		zap.String("topic", c.Reader.Config().Topic))
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	var messages []models.ScanResultMessage
	total := 0

	for len(messages) < batchSize {
		msg, err := c.Reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				c.logger.Info("scan-result-reducer-consumer: таймаут, заканчиваем чтение")
				break
			}
			return nil, total, err
		}

		var resultMsg models.ScanResultMessage
		if err := json.Unmarshal(msg.Value, &resultMsg); err != nil {
			c.logger.Error("scan-result-reducer-consumer: unmarshal error", zap.Error(err))
			// не прибавляем total — пропускаем
			continue
		}

		messages = append(messages, resultMsg)
		total++
		c.logger.Info("scan-result-reducer-consumer: Получено сообщение", zap.Any("resultMsg", resultMsg))

		if err := c.Reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("scan-result-reducer-consumer: Ошибка commit", zap.Error(err))
		}
	}

	c.logger.Info("scan-result-reducer-consumer: чтение сообщений из Kafka из топика завершено")
	return messages, total, nil
}

func (c *ScanResultConsumer) CloseReader() error {
	c.logger.Info("scan-result-reducer-consumer: Закрытие Kafka-консьюмера")
	return c.Reader.Close()
}
