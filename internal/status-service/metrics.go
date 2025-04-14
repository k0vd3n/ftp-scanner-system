package statusservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Вектор и обычная метрика для количества полученных запросов на получение статуса
	requestsTotalVec *prometheus.CounterVec
	RequestsTotal    prometheus.Counter

	// Вектор и обычная метрика для гистограммы времени обработки запроса
	processingDurationVec *prometheus.HistogramVec
	ProcessingDuration    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени выполнения запроса к базе данных
	dbQueryDurationVec *prometheus.HistogramVec
	DbQueryDuration    prometheus.Histogram

	// Вектор и обычная метрика для количества ошибок при обработке запроса
	errorsTotalVec *prometheus.CounterVec
	ErrorsTotal    prometheus.Counter
)

// InitMetrics регистрирует метрики с лейблом instance и инициализирует их
func InitMetrics(instance string) {
	// Инициализация векторов метрик
	requestsTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "status_service_requests_total",
			Help: "Общее количество запросов к статус-сервису",
		},
		[]string{"instance"},
	)
	processingDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "status_service_processing_duration_seconds",
			Help:    "Время обработки запроса на получение статуса",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"instance"},
	)
	dbQueryDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "status_service_db_query_duration_seconds",
			Help:    "Время выполнения запроса к базе данных для получения счетчиков",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"instance"},
	)
	errorsTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "status_service_errors_total",
			Help: "Количество ошибок, возникших при обработке запросов в статус-сервисе",
		},
		[]string{"instance"},
	)

	// Регистрация векторов метрик
	prometheus.MustRegister(
		requestsTotalVec,
		processingDurationVec,
		dbQueryDurationVec,
		errorsTotalVec,
	)

	// Инициализация обычных метрик с привязкой лейбла instance
	RequestsTotal = requestsTotalVec.WithLabelValues(instance)
	ProcessingDuration = processingDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	DbQueryDuration = dbQueryDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ErrorsTotal = errorsTotalVec.WithLabelValues(instance)
}
