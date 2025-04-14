package reportservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Вектор и обычная метрика для количества запросов на генерацию отчёта
	reportRequestsTotalVec *prometheus.CounterVec
	ReportRequestsTotal    prometheus.Counter

	// Вектор и обычная метрика для гистограммы времени обработки запроса
	reportProcessingDurationVec *prometheus.HistogramVec
	ReportProcessingDuration    prometheus.Histogram

	// Вектор и обычная метрика для количества ошибок при генерации отчёта
	reportErrorsTotalVec *prometheus.CounterVec
	ReportErrorsTotal    prometheus.Counter

	// Вектор и обычная метрика для времени сохранения отчёта
	reportStorageDurationVec *prometheus.HistogramVec
	ReportStorageDuration    prometheus.Histogram
)

// InitMetrics регистрирует метрики с лейблом instance и инициализирует их
func InitMetrics(instance string) {
	// Инициализация векторов метрик
	reportRequestsTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "report_service_requests_total",
			Help: "Общее количество запросов на генерацию отчёта",
		},
		[]string{"instance"},
	)
	reportProcessingDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "report_service_processing_duration_seconds",
			Help:    "Время обработки запроса генерации отчёта (от запроса до ответа)",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"instance"},
	)
	reportErrorsTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "report_service_errors_total",
			Help: "Количество ошибок при генерации отчёта",
		},
		[]string{"instance"},
	)
	reportStorageDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "report_service_storage_duration_seconds",
			Help:    "Время сохранения отчёта в хранилище (файловая система)",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"instance"},
	)

	// Регистрация векторов метрик
	prometheus.MustRegister(
		reportRequestsTotalVec,
		reportProcessingDurationVec,
		reportErrorsTotalVec,
		reportStorageDurationVec,
	)

	// Инициализация обычных метрик с привязкой лейбла instance
	ReportRequestsTotal = reportRequestsTotalVec.WithLabelValues(instance)
	ReportProcessingDuration = reportProcessingDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ReportErrorsTotal = reportErrorsTotalVec.WithLabelValues(instance)
	ReportStorageDuration = reportStorageDurationVec.WithLabelValues(instance).(prometheus.Histogram)
}
