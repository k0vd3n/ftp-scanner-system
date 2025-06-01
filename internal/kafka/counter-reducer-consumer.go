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

type CounterConsumer struct {
	Reader *kafka.Reader
	logger *zap.Logger
}

func NewCounterConsumer(brokers []string, topic, groupID string, logger *zap.Logger) KafkaCounterReducerConsumerInterface {
	return &CounterConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			CommitInterval: 0,
		}),
		logger: logger,
	}
}

func (c *CounterConsumer) ReadMessages(
	parentCtx context.Context,
	batchSize int,
	duration time.Duration,
) ([]models.CountMessage, error) {
	c.logger.Info("counter-reducer-consumer: чтение сообщений из Kafka из топика", zap.String("topic", c.Reader.Config().Topic))

	// таймаутный контекст
	ctx, cancel := context.WithTimeout(parentCtx, duration)
	defer cancel()

	var msgs []models.CountMessage
	total := 0

	for len(msgs) < batchSize {
		kafkaMsg, err := c.Reader.FetchMessage(ctx)
		if err != nil {
			// таймаут или контекст истёк
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				c.logger.Info("counter-reducer-consumer: таймаут, заканчиваем чтение")
				break
			}
			return nil, err
		}

		var resultMsg models.CountMessage
		if err := json.Unmarshal(kafkaMsg.Value, &resultMsg); err != nil {
			c.logger.Error("counter-reducer-consumer: unmarshal error", zap.Error(err))
			continue
		}

		msgs = append(msgs, resultMsg)
		total++
		c.logger.Info("counter-reducer-consumer: получено сообщение", zap.Any("resultMsg", resultMsg))

		if err := c.Reader.CommitMessages(ctx, kafkaMsg); err != nil {
			c.logger.Error("counter-reducer-consumer: commit error", zap.Error(err))
		}
	}

	c.logger.Info("counter-reducer-consumer: чтение завершено", zap.Int("messages", len(msgs)))
	return msgs, nil
}

func (c *CounterConsumer) CloseReader() error {
	c.logger.Info("counter-reducer-consumer: закрытие Kafka-консьюмера")
	return c.Reader.Close()
}
