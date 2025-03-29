package mainservice

import (
	"log"
	"time"

	"ftp-scanner_try2/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Количество входящих HTTP-запросов к main-service
	RequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "main_service_counter",
		Help: "Общее количество HTTP запросов, полученных main-service",
	})

	// Гистограмма времени обработки HTTP-запроса в main-service
	ProcessingScanRequest = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "main_service_processing_scan_request_duration_seconds",
		Help:    "Время обработки HTTP запроса main-service",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	ProcessingStartScanMethod = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "main_service_start_scan_method_duration_seconds",
		Help:    "Время запуска метода НачатьСканирование",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	// Гистограмма времени публикации Kafka-сообщения (при запуске сканирования)
	KafkaPublishDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "main_service_kafka_publish_duration_seconds",
		Help:    "Время публикации сообщения в Kafka для запуска сканирования",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	ProcessingReportRequest = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "main_service_processing_report_request_duration_seconds",
		Help:    "Время обработки HTTP запроса main-service",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	// Гистограмма времени выполнения gRPC вызова к Report Service
	GRPCReportCallDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "main_service_grpc_report_call_duration_seconds",
		Help:    "Время выполнения gRPC вызова к Report Service",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	ProcessingStatusRequest = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "main_service_processing_status_request_duration_seconds",
		Help:    "Время получения статуса сканирования",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
	// Гистограмма времени выполнения gRPC вызова к Status Service
	GRPCStatusCallDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "main_service_grpc_status_call_duration_seconds",
		Help:    "Время выполнения gRPC вызова к Status Service",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	// Счётчик ошибок в main-service
	ErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "main_service_errors_total_counter",
		Help: "Общее количество ошибок, произошедших в main-service",
	})
)

func InitMetrics() {
	prometheus.MustRegister(
		RequestsTotal,
		ProcessingScanRequest,
		ProcessingStartScanMethod,
		KafkaPublishDuration,
		ProcessingReportRequest,
		GRPCReportCallDuration,
		ProcessingStatusRequest,
		GRPCStatusCallDuration,
		ErrorsTotal,
	)
}

func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(RequestsTotal).
		Collector(ProcessingScanRequest).
		Collector(ProcessingStartScanMethod).
		Collector(KafkaPublishDuration).
		Collector(ProcessingReportRequest).
		Collector(GRPCReportCallDuration).
		Collector(ProcessingStatusRequest).
		Collector(GRPCStatusCallDuration).
		Collector(ErrorsTotal).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("main-service metrics: Ошибка отправки метрик в Pushgateway: %v", err)
	} else {

		log.Printf("main-service metrics: Метрики отправлены в Pushgateway")
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
