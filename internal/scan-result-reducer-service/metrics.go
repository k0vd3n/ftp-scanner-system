package scanresultreducerservice

import (
	"ftp-scanner_try2/config"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Количество полученных сообщений (батчей) для редьюса
	ReceivedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "scan_result_reducer_received_messages_total",
		Help: "Общее количество полученных сообщений для редьюса результатов сканирования",
	})

	// Гистограмма времени обработки пакета сообщений
	ProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "scan_result_reducer_processing_duration_seconds",
		Help:    "Время редьюсирования и записи отчётов в MongoDB",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
	})

	// Количество сформированных отчётов (уникальных сканов)
	ReportsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "scan_result_reducer_reports_total",
		Help: "Общее количество сформированных отчётов после редьюса",
	})

	// Счётчик ошибок (например, ошибок записи в MongoDB)
	ErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "scan_result_reducer_errors_total",
		Help: "Количество ошибок при редьюсе или записи данных в MongoDB",
	})
)

func InitMetrics() {
	prometheus.MustRegister(
		ReceivedMessages,
		ProcessingDuration,
		ReportsTotal,
		ErrorsTotal,
	)
}

// PushMetrics отправляет метрики в Pushgateway.
// cfg передаётся из конфигурации (включая URL, имя job, instance и интервал пуша).
func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(ReceivedMessages).
		Collector(ProcessingDuration).
		Collector(ReportsTotal).
		Collector(ErrorsTotal).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("scan-result-reducer-service metrics: Ошибка при отправке метрик в Pushgateway: %v", err)
	} else {
		log.Printf("scan-result-reducer-service metrics: Метрики отправлены в Pushgateway")
	}
}

// StartPushLoop запускает периодическую отправку метрик в Pushgateway.
// Здесь используется поле cfg.PushInterval, которое можно указать в конфигурации (в секундах).
func StartPushLoop(cfg *config.PushGatewayConfig) {
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.PushInterval) * time.Second)
		for range ticker.C {
			PushMetrics(cfg)
		}
	}()
}
