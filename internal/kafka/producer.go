package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewProducer(broker string, logger *zap.Logger) (KafkaPoducerInterface, error) {
	writer := &kafka.Writer{
		Addr: kafka.TCP(broker),
		// Topic: topic, // Топик по умолчанию
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{
		writer: writer,
		logger: logger,
	}, nil
}

func (p *Producer) SendMessage(topic string, message interface{}) error {
	p.logger.Info("Producer: Отправка сообщения в Kafka",
		zap.String("topic", topic))
	// Сериализация сообщения в JSON
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Отправка сообщения в Kafka
	err = p.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: topic,
		Value: msgBytes,
	})
	if err != nil {
		p.logger.Error("Producer: Ошибка отправки сообщения", zap.Error(err))

		return err
	}

	p.logger.Info("Producer: Сообщение отправлено в Kafka",
		zap.String("topic", topic),
		zap.Any("message", message))
	return nil
}

func (p *Producer) CloseWriter() error {
	p.logger.Info("Producer: Закрытие Kafka-продюсера")
	return p.writer.Close()
}
