package scanresultreducerservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Вектор и обычная метрика для количества полученных сообщений (батчей) для редьюса
	receivedMessagesVec *prometheus.CounterVec
	ReceivedMessages    prometheus.Counter

	receivedMessagesKafkaHandlerVec          *prometheus.CounterVec
	ReceivedMessagesKafkaHandler prometheus.Counter

	// Вектор и обычная метрика для количества полученных файлов для редьюсирования
	receivedFilesVec *prometheus.CounterVec
	ReceivedFiles    prometheus.Counter

	// Вектор и обычная метрика для количества полученных директорий для редьюсирования
	receivedDirectoriesVec *prometheus.CounterVec
	ReceivedDirectories    prometheus.Counter

	// Вектор и обычная метрика для гистограммы времени обработки пакета сообщений
	processingDurationVec *prometheus.HistogramVec
	ProcessingDuration    prometheus.Histogram

	// Вектор и обычная метрика для количества сформированных отчётов (уникальных сканов)
	reportsTotalVec *prometheus.CounterVec
	ReportsTotal    prometheus.Counter

	// Вектор и обычная метрика для счётчика ошибок (например, ошибок записи в MongoDB)
	errorsTotalVec *prometheus.CounterVec
	ErrorsTotal    prometheus.Counter
)

// InitMetrics регистрирует метрики с лейблом instance и инициализирует их
func InitMetrics(instance string) {
	// Инициализация векторов метрик
	receivedMessagesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scan_result_reducer_received_messages_total",
			Help: "Общее количество полученных сообщений для редьюса результатов сканирования",
		},
		[]string{"instance"},
	)

	receivedMessagesKafkaHandlerVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scan_result_reducer_received_messages_kafka_handler_total",
			Help: "Общее количество полученных сообщений для редьюса результатов сканирования",
		},
		[]string{"instance"},
	)
	receivedFilesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scan_result_reducer_received_numbers_total",
			Help: "Общее количество полученных чисел для редьюса результатов сканирования",
		},
		[]string{"instance"},
	)
	receivedDirectoriesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scan_result_reducer_received_directories_total",
			Help: "Общее количество полученных директорий для редьюса результатов сканирования",
		},
		[]string{"instance"},
	)
	processingDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scan_result_reducer_processing_duration_seconds",
			Help:    "Время редьюсирования и записи отчётов в MongoDB",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"instance"},
	)
	reportsTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scan_result_reducer_reports_total",
			Help: "Общее количество сформированных отчётов после редьюса",
		},
		[]string{"instance"},
	)
	errorsTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scan_result_reducer_errors_total",
			Help: "Количество ошибок при редьюсе или записи данных в MongoDB",
		},
		[]string{"instance"},
	)

	// Регистрация векторов метрик
	prometheus.MustRegister(
		receivedMessagesVec,
		receivedMessagesKafkaHandlerVec,
		receivedFilesVec,
		receivedDirectoriesVec,
		processingDurationVec,
		reportsTotalVec,
		errorsTotalVec,
	)

	// Инициализация обычных метрик с привязкой лейбла instance
	ReceivedMessages = receivedMessagesVec.WithLabelValues(instance)
	ReceivedMessagesKafkaHandler = receivedMessagesKafkaHandlerVec.WithLabelValues(instance)
	ReceivedFiles = receivedFilesVec.WithLabelValues(instance)
	ReceivedDirectories = receivedDirectoriesVec.WithLabelValues(instance)
	ProcessingDuration = processingDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ReportsTotal = reportsTotalVec.WithLabelValues(instance)
	ErrorsTotal = errorsTotalVec.WithLabelValues(instance)
}
