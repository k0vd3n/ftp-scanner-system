package counterreducerservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Вектор и обычная метрика для счётчика полученных сообщений (batch)
	receivedMessagesVec *prometheus.CounterVec
	ReceivedMessages    prometheus.Counter

	// Вектор и обычная метрика для счётчика числа полученных чисел для редьюсирования
	receivedNumbersVec *prometheus.CounterVec
	ReceivedNumbers    prometheus.Counter

	// Вектор и обычная метрика для гистограммы времени редьюсирования
	processingReduceVec *prometheus.HistogramVec
	ProcessingReduce    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени записи в MongoDB
	processingInsertVec *prometheus.HistogramVec
	ProcessingInsert    prometheus.Histogram

	// Вектор и обычная метрика для гистограммы времени редьюса и записи в MongoDB
	processingDurationVec *prometheus.HistogramVec
	ProcessingDuration    prometheus.Histogram

	// Вектор и обычная метрика для счётчика редьюсированных сообщений (уникальных)
	reducedMessagesVec *prometheus.CounterVec
	ReducedMessages    prometheus.Counter

	// Вектор и обычная метрика для счётчика ошибок вставки в MongoDB
	mongoInsertionErrorsVec *prometheus.CounterVec
	MongoInsertionErrors    prometheus.Counter
)

// InitMetrics регистрирует метрики с лейблом instance и инициализирует их
func InitMetrics(instance string) {
	// Инициализация векторов метрик
	receivedMessagesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_reducer_received_messages_total",
			Help: "Количество полученных сообщений для редьюсирования",
		},
		[]string{"instance"},
	)

	receivedNumbersVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_reducer_received_numbers_total",
			Help: "Количество полученных чисел для редьюсирования",
		},
		[]string{"instance"},
	)

	processingReduceVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "counter_reducer_processing_reduce_duration_seconds",
			Help:    "Время редьюсирования",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"instance"},
	)
	processingInsertVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "counter_reducer_processing_insert_duration_seconds",
			Help:    "Время записи в MongoDB",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"instance"},
	)
	processingDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "counter_reducer_processing_duration_seconds",
			Help:    "Время редьюса и записи в MongoDB",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"instance"},
	)
	reducedMessagesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_reducer_reduced_messages_total",
			Help: "Количество редьюсированных сообщений",
		},
		[]string{"instance"},
	)
	mongoInsertionErrorsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_reducer_mongo_insertion_errors_total",
			Help: "Количество ошибок при вставке редьюсированных данных в MongoDB",
		},
		[]string{"instance"},
	)

	// Регистрация векторов метрик
	prometheus.MustRegister(
		receivedMessagesVec,
		receivedNumbersVec,
		processingReduceVec,
		processingInsertVec,
		processingDurationVec,
		reducedMessagesVec,
		mongoInsertionErrorsVec,
	)

	// Инициализация метрик с лейблом instance
	ReceivedMessages = receivedMessagesVec.WithLabelValues(instance)
	ReceivedNumbers = receivedNumbersVec.WithLabelValues(instance)
	ProcessingReduce = processingReduceVec.WithLabelValues(instance).(prometheus.Histogram)
	ProcessingInsert = processingInsertVec.WithLabelValues(instance).(prometheus.Histogram)
	ProcessingDuration = processingDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ReducedMessages = reducedMessagesVec.WithLabelValues(instance)
	MongoInsertionErrors = mongoInsertionErrorsVec.WithLabelValues(instance)
}
