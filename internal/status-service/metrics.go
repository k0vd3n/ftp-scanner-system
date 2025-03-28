package statusservice

import (
	"log"
	"time"

	"ftp-scanner_try2/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Количество полученных запросов на получение статуса
	RequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "status_service_requests_total",
		Help: "Общее количество запросов к статус-сервису",
	})
	// Гистограмма времени обработки запроса (от входа до отправки ответа)
	ProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "status_service_processing_duration_seconds",
		Help:    "Время обработки запроса на получение статуса",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
	// Гистограмма времени выполнения запроса к базе (например, для получения счетчиков)
	DbQueryDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "status_service_db_query_duration_seconds",
		Help:    "Время выполнения запроса к базе данных для получения счетчиков",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
	// Количество ошибок при обработке запроса
	ErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "status_service_errors_total",
		Help: "Количество ошибок, возникших при обработке запросов в статус-сервисе",
	})
)

func InitMetrics() {
	prometheus.MustRegister(
		RequestsTotal,
		ProcessingDuration,
		DbQueryDuration,
		ErrorsTotal,
	)
}

// PushMetrics отправляет метрики в Pushgateway
func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(RequestsTotal).
		Collector(ProcessingDuration).
		Collector(DbQueryDuration).
		Collector(ErrorsTotal).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("status-service metrics: Ошибка при отправке метрик в Pushgateway: %v", err)
	} else {
		log.Printf("status-service metrics: Метрики отправлены в Pushgateway")
	}
}

// StartPushLoop запускает периодическую отправку метрик в Pushgateway.
// В конфигурации должен быть указан параметр PushInterval (в секундах).
func StartPushLoop(cfg *config.PushGatewayConfig) {
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.PushInterval) * time.Second)
		for range ticker.C {
			PushMetrics(cfg)
		}
	}()
}
