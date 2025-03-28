package counterreducerservice

import (
	"ftp-scanner_try2/config"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Счётчик полученных сообщений (batch)
	ReceivedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "counter_reducer_received_messages_total",
		Help: "Количество полученных сообщений для редьюсирования",
	})
	ProcessingReduce = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "counter_reducer_processing_reduce_duration_seconds",
		Help:    "Время редьюсирования",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
	})
	ProcessingInsert = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "counter_reducer_processing_insert_duration_seconds",
		Help:    "Время записи в MongoDB",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
	})
	ProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "counter_reducer_processing_duration_seconds",
		Help:    "Время редьюса и записи в MongoDB",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
	})
	// Счётчик редьюсированных сканов (уникальных)
	ReducedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "counter_reducer_reduced_messages_total",
		Help: "Количество редьюсированных сообщений",
	})
	// Счётчик ошибок записи в MongoDB
	MongoInsertionErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "counter_reducer_mongo_insertion_errors_total",
		Help: "Количество ошибок при вставке редьюсированных данных в MongoDB",
	})
)

func InitMetrics() {
	prometheus.MustRegister(
		ReceivedMessages,
		ProcessingReduce,
		ProcessingInsert,
		ProcessingDuration,
		ReducedMessages,
		MongoInsertionErrors,
	)
}

func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(ReceivedMessages).
		Collector(ProcessingReduce).
		Collector(ProcessingInsert).
		Collector(ProcessingDuration).
		Collector(ReducedMessages).
		Collector(MongoInsertionErrors).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("counter reducer service metrics: Ошибка при отправке метрик в Pushgateway: %v", err)
	} else {
		log.Printf("counter reducer service metrics: Метрики отправлены в Pushgateway")
	}
}

func StartPushLoop(cfg *config.PushGatewayConfig) {
	go func() {
		// Если в конфиге указан интервал (например, в секундах), его можно использовать
		ticker := time.NewTicker(time.Duration(cfg.PushInterval) * time.Second)
		for range ticker.C {
			PushMetrics(cfg)
		}
	}()
}
