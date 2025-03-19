package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"ftp-scanner_try2/config"
	"ftp-scanner_try2/internal/models"
	"math"
	"strconv"
	"strings"

	"github.com/segmentio/kafka-go"
)

type KafkaScanResultProducer struct {
	writer *kafka.Writer
	// topics       config.SizeBasedRouterTopics
	routingRules config.RoutingConfig
}

func NewScanResultProducer(
	broker string,
	routing config.RoutingConfig,
	// topics config.SizeBasedRouterTopics,
) (*KafkaScanResultProducer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 0,
	}

	return &KafkaScanResultProducer{
		writer: writer,
	}, nil
}

func (k *KafkaScanResultProducer) SendMessage(message models.ScanResultMessage) error {
	topics := k.determineTopics(message)
	for _, topic := range topics {
		err := k.sendToTopic(topic, message)
		if err != nil {
			return fmt.Errorf("failed to send message to topic %s: %v", topic, err)
		}
	}
	return nil
}

func (k *KafkaScanResultProducer) CloseWriter() error {
	return k.writer.Close()
}

func (k *KafkaScanResultProducer) determineTopics(msg models.ScanResultMessage) []string {
	// Всегда добавляем топик по умолчанию
	topics := []string{k.routingRules.DefaultTopic}

	// Ищем подходящие правила для типа сканирования
	for _, rule := range k.routingRules.Rules {
		if rule.ScanType != msg.ScanType {
			continue
		}

		// Проверяем условие срабатывания правила
		if k.isRuleTriggered(rule, msg.Result) {
			topics = append(topics, rule.OutputTopics...)
		}
	}

	return topics
}

func (k *KafkaScanResultProducer) isRuleTriggered(rule config.RoutingRule, result string) bool {
	resInt, err := strconv.Atoi(result)
	if err != nil {
		return false
	}

	parts := strings.Split(rule.TriggerValue, "-")
	switch len(parts) {
	case 2: // Диапазон
		min, max := 0, math.MaxInt32

		if parts[0] != "" {
			min, _ = strconv.Atoi(parts[0])
		}
		if parts[1] != "" {
			max, _ = strconv.Atoi(parts[1])
		}

		return resInt >= min && (max == math.MaxInt32 || resInt < max)

	default: // Простое значение
		return result == rule.TriggerValue
	}
}

func (k *KafkaScanResultProducer) sendToTopic(topic string, msg models.ScanResultMessage) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return k.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: topic,
		Value: jsonMsg,
	})
}
