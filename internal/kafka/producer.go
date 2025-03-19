package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) (KafkaPoducerInterface, error) {
	writer := &kafka.Writer{
		Addr:  kafka.TCP(broker),
		// Topic: topic, // Топик по умолчанию
		Balancer: &kafka.LeastBytes{},
		BatchSize: 1,
		BatchTimeout: 0,
	}

	return &Producer{
		writer: writer,
	}, nil
}

func (p *Producer) SendMessage(topic string, message interface{}) error {
	log.Printf("Producer: Отправка сообщения в Kafka...")
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
		log.Printf("Producer: Ошибка при отправке сообщения в Kafka: %v", err)
		return err
	}

	log.Printf("Producer: Сообщение отправлено в топик %s: %v", topic, message)
	return nil
}

func (p *Producer) CloseWriter() error {
	log.Println("Producer: Закрытие Kafka-продусера...")
	return p.writer.Close()
}
