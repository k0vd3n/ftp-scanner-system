package mainservice

import (
	"log"
	"time"

	"ftp-scanner_try2/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Вектор и обычная метрика для количества входящих HTTP-запросов
	requestsTotalVec *prometheus.CounterVec
	RequestsTotal    prometheus.Counter

	// Вектор и обычная метрика для гистограммы времени обработки HTTP-запроса
	processingScanRequestVec *prometheus.HistogramVec
	ProcessingScanRequest    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени запуска метода "НачатьСканирование"
	processingStartScanMethodVec *prometheus.HistogramVec
	ProcessingStartScanMethod    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени публикации сообщения в Kafka
	kafkaPublishDurationVec *prometheus.HistogramVec
	KafkaPublishDuration    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени обработки запроса отчёта
	processingReportRequestVec *prometheus.HistogramVec
	ProcessingReportRequest    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени выполнения gRPC вызова к Report Service
	grpcReportCallDurationVec *prometheus.HistogramVec
	GRPCReportCallDuration    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени получения статуса сканирования
	processingStatusRequestVec *prometheus.HistogramVec
	ProcessingStatusRequest    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени выполнения gRPC вызова к Status Service
	grpcStatusCallDurationVec *prometheus.HistogramVec
	GRPCStatusCallDuration    prometheus.Histogram

	// Вектор и обычная метрика для счётчика ошибок
	errorsTotalVec *prometheus.CounterVec
	ErrorsTotal    prometheus.Counter
)

// InitMetrics регистрирует все метрики с лейблом instance и инициализирует их
func InitMetrics(instance string) {
	// Инициализация векторов метрик
	requestsTotalVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "main_service_counter",
		Help: "Общее количество HTTP запросов, полученных main-service",
	}, []string{"instance"})

	processingScanRequestVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "main_service_processing_scan_request_duration_seconds",
		Help:    "Время обработки HTTP запроса main-service",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"instance"})

	processingStartScanMethodVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "main_service_start_scan_method_duration_seconds",
		Help:    "Время запуска метода НачатьСканирование",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"instance"})

	kafkaPublishDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "main_service_kafka_publish_duration_seconds",
		Help:    "Время публикации сообщения в Kafka для запуска сканирования",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"instance"})

	processingReportRequestVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "main_service_processing_report_request_duration_seconds",
		Help:    "Время обработки HTTP запроса main-service (report)",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"instance"})

	grpcReportCallDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "main_service_grpc_report_call_duration_seconds",
		Help:    "Время выполнения gRPC вызова к Report Service",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"instance"})

	processingStatusRequestVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "main_service_processing_status_request_duration_seconds",
		Help:    "Время получения статуса сканирования",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"instance"})

	grpcStatusCallDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "main_service_grpc_status_call_duration_seconds",
		Help:    "Время выполнения gRPC вызова к Status Service",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"instance"})

	errorsTotalVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "main_service_errors_total_counter",
		Help: "Общее количество ошибок, произошедших в main-service",
	}, []string{"instance"})

	// Регистрация всех векторов метрик
	prometheus.MustRegister(
		requestsTotalVec,
		processingScanRequestVec,
		processingStartScanMethodVec,
		kafkaPublishDurationVec,
		processingReportRequestVec,
		grpcReportCallDurationVec,
		processingStatusRequestVec,
		grpcStatusCallDurationVec,
		errorsTotalVec,
	)

	// Инициализация обычных метрик с привязкой лейбла instance
	RequestsTotal = requestsTotalVec.WithLabelValues(instance)
	ProcessingScanRequest = processingScanRequestVec.WithLabelValues(instance).(prometheus.Histogram)
	ProcessingStartScanMethod = processingStartScanMethodVec.WithLabelValues(instance).(prometheus.Histogram)
	KafkaPublishDuration = kafkaPublishDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ProcessingReportRequest = processingReportRequestVec.WithLabelValues(instance).(prometheus.Histogram)
	GRPCReportCallDuration = grpcReportCallDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ProcessingStatusRequest = processingStatusRequestVec.WithLabelValues(instance).(prometheus.Histogram)
	GRPCStatusCallDuration = grpcStatusCallDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ErrorsTotal = errorsTotalVec.WithLabelValues(instance)
}

// PushMetrics отправляет метрики в Pushgateway
func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(requestsTotalVec).
		Collector(processingScanRequestVec).
		Collector(processingStartScanMethodVec).
		Collector(kafkaPublishDurationVec).
		Collector(processingReportRequestVec).
		Collector(grpcReportCallDurationVec).
		Collector(processingStatusRequestVec).
		Collector(grpcStatusCallDurationVec).
		Collector(errorsTotalVec).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("main-service metrics: Ошибка отправки метрик в Pushgateway: %v", err)
	} else {
		log.Printf("main-service metrics: Метрики отправлены в Pushgateway")
	}
}

// StartPushLoop запускает периодическую отправку метрик в Pushgateway
func StartPushLoop(cfg *config.PushGatewayConfig) {
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.PushInterval) * time.Second)
		for range ticker.C {
			PushMetrics(cfg)
		}
	}()
}
