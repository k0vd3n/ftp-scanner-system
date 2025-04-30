package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"ftp-scanner_try2/internal/models"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type CounterConsumer struct {
	Reader *kafka.Reader
}

func NewCounterConsumer(brokers []string, topic, groupID string) KafkaCounterReducerConsumerInterface {
	return &CounterConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			CommitInterval: 0,
		}),
	}
}

func (c *CounterConsumer) ReadMessages(
	parentCtx context.Context,
	batchSize int,
	duration time.Duration,
) ([]models.CountMessage, error) {
	log.Printf("counter-reducer-consumer: читаем из топика %s...", c.Reader.Config().Topic)

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
				log.Println("counter-reducer-consumer: таймаут, заканчиваем чтение")
				break
			}
			return nil, err
		}

		var resultMsg models.CountMessage
		if err := json.Unmarshal(kafkaMsg.Value, &resultMsg); err != nil {
			log.Printf("counter-reducer-consumer: unmarshal error: %v", err)
			continue
		}


		msgs = append(msgs, resultMsg)
		total++
		log.Printf("counter-reducer-consumer: получено сообщение: %s", string(kafkaMsg.Value))
		
		if err := c.Reader.CommitMessages(ctx, kafkaMsg); err != nil {
			log.Printf("counter-reducer-consumer: Ошибка commit: %v", err)
		}
	}

	log.Printf("counter-reducer-consumer: всего сообщений: %d", total)
	return msgs, nil
}

func (c *CounterConsumer) CloseReader() error {
	log.Println("Counter-reducer-consumer: Закрытие Kafka-консьюмера...")
	return c.Reader.Close()
}
