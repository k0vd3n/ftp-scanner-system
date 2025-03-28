package reportservice

import (
	"log"
	"time"

	"ftp-scanner_try2/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Количество полученных запросов на генерацию отчёта
	ReportRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "report_service_counter",
		Help: "Общее количество запросов на генерацию отчёта",
	})
	// Гистограмма времени от получения запроса до выдачи ответа
	ReportProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "report_service_histogram",
		Help:    "Время обработки запроса генерации отчёта (от запроса до ответа)",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
	// Счётчик ошибок при генерации отчёта
	ReportErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "report_service_counter",
		Help: "Количество ошибок при генерации отчёта",
	})
	// Гистограмма времени сохранения отчёта в хранилище
	ReportStorageDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "report_service_histogram",
		Help:    "Время сохранения отчёта в хранилище (файловая система)",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
)

func InitMetrics() {
	prometheus.MustRegister(
		ReportRequestsTotal,
		ReportProcessingDuration,
		ReportErrorsTotal,
		ReportStorageDuration,
	)
}

func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(ReportRequestsTotal).
		Collector(ReportProcessingDuration).
		Collector(ReportErrorsTotal).
		Collector(ReportStorageDuration).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("report-service metrics: Ошибка при отправке метрик в Pushgateway: %v", err)
	} else {
		log.Printf("report-service metrics: Метрики отправлены в Pushgateway")
	}
}

func StartPushLoop(cfg *config.PushGatewayConfig) {
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.PushInterval) * time.Second)
		for range ticker.C {
			PushMetrics(cfg)
		}
	}()
}
