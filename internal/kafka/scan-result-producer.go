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
	"go.uber.org/zap"
)

type KafkaScanResultProducer struct {
	writer       *kafka.Writer
	routingRules config.RoutingConfig
	logger       *zap.Logger
}

func NewScanResultProducer(
	broker string,
	routing config.RoutingConfig,
	logger *zap.Logger,
) (KafkaScanResultPoducerInterface, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10,
	}

	return &KafkaScanResultProducer{
		writer:       writer,
		routingRules: routing,
		logger:       logger,
	}, nil
}

func (k *KafkaScanResultProducer) SendMessage(message models.ScanResultMessage) error {
	k.logger.Info("Scan-result-producer: Отправка сообщения в Kafka", zap.Any("message", message))
	topics := k.determineTopics(message)
	k.logger.Info("Scan-result-producer: Топики", zap.Any("topics", topics))
	for _, topic := range topics {
		err := k.sendToTopic(topic, message)
		if err != nil {
			k.logger.Error("Scan-result-producer: Ошибка при отправке сообщения в Kafka", zap.Error(err))
			return fmt.Errorf("scan-result-producer: Ошибка при отправке сообщения в Kafka %s: %v", topic, err)
		}
		k.logger.Info("Scan-result-producer: Сообщение отправлено в топик",
			zap.String("topic", topic),
			zap.Any("message", message))
	}
	k.logger.Info("Scan-result-producer: Сообщения отправлены в топики",
		zap.Any("topics", topics),
		zap.Any("message", message))
	return nil
}

func (k *KafkaScanResultProducer) CloseWriter() error {
	k.logger.Info("Scan-result-producer: Закрытие Kafka-продусера")
	return k.writer.Close()
}

func (k *KafkaScanResultProducer) determineTopics(msg models.ScanResultMessage) []string {
	k.logger.Info("Scan-result-producer: Определение топиков", zap.Any("msg", msg))
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
	// Попробуем как число (для zero_bytes, line_count и т.п.)
	resInt, err := strconv.Atoi(result)
	if err == nil {
		// Число: проверка диапазона или точного совпадения
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
		default: // Точное значение
			return result == rule.TriggerValue
		}
	}

	// Если не число — проверка MIME или строки (напр. file_extension)
	// Поддержим wildcard: если trigger_value заканчивается на "/"
	if strings.HasSuffix(rule.TriggerValue, "/") {
		return strings.HasPrefix(result, rule.TriggerValue)
	}

	// Строгое строковое сравнение
	return result == rule.TriggerValue
}

func (k *KafkaScanResultProducer) sendToTopic(topic string, msg models.ScanResultMessage) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		k.logger.Error("Scan-result-producer: Ошибка при сериализации сообщения в JSON", zap.Error(err))
		return err
	}

	return k.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: topic,
		Value: jsonMsg,
	})
}
