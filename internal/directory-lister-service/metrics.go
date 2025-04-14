package directorylisterservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	receivedMessagesVec          *prometheus.CounterVec
	ReceivedMessages             prometheus.Counter
	readErrorsVec                *prometheus.CounterVec
	ReadErrors                   prometheus.Counter
	processingDurationVec        *prometheus.HistogramVec
	ProcessingDuration           prometheus.Histogram
	ftpReconnectionsVec          *prometheus.CounterVec
	FtpReconnections             prometheus.Counter
	directoriesProcessedVec      *prometheus.CounterVec
	DirectoriesProcessed         prometheus.Counter
	filesFoundVec                *prometheus.CounterVec
	FilesFound                   prometheus.Counter
	ftpListDurationVec           *prometheus.HistogramVec
	FtpListDuration              prometheus.Histogram
	kafkaMessagesSentDirsVec     *prometheus.CounterVec
	KafkaMessagesSentDirectories prometheus.Counter
	kafkaMessagesSentFilesVec    *prometheus.CounterVec
	KafkaMessagesSentFiles       prometheus.Counter
	kafkaMessagesReturnedDirsVec *prometheus.CounterVec
	KafkaMessagesReturnedDirs    prometheus.Counter

	kafkaMessagesSentScanFilesCountVec *prometheus.CounterVec
	KafkaMessagesSentScanFilesCount    prometheus.Counter

	kafkaMessagesSentScanDirsCountVec *prometheus.CounterVec
	KafkaMessagesSentScanDirsCount    prometheus.Counter

	kafkaMessagesSentCompletedDirsCountVec *prometheus.CounterVec
	KafkaMessagesSentCompletedDirsCount    prometheus.Counter

	kafkaSendErrorsVec           *prometheus.CounterVec
	KafkaSendErrors              prometheus.Counter
)

// InitMetrics регистрирует метрики с лейблом instance
func InitMetrics(instance string) {
	// Инициализация векторов метрик
	receivedMessagesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_received_messages_total",
			Help: "Количество успешно полученных сообщений из Kafka",
		},
		[]string{"instance"},
	)
	readErrorsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_read_errors_total",
			Help: "Количество ошибок при чтении сообщений из Kafka",
		},
		[]string{"instance"},
	)
	processingDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "directory_lister_processing_duration_seconds",
			Help:    "Время обработки одного сообщения",
			Buckets: prometheus.ExponentialBuckets(0.00005, 2, 20),
		},
		[]string{"instance"},
	)
	ftpReconnectionsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_ftp_reconnections_total",
			Help: "Количество событий переподключения к FTP",
		},
		[]string{"instance"},
	)
	directoriesProcessedVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_directories_processed_total",
			Help: "Количество обработанных директорий",
		},
		[]string{"instance"},
	)
	filesFoundVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_files_found_total",
			Help: "Количество найденных файлов",
		},
		[]string{"instance"},
	)
	ftpListDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "directory_lister_ftp_list_directory_duration_seconds",
			Help:    "Время листинга директории",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"instance"},
	)
	kafkaMessagesSentDirsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_kafka_messages_sent_directories_total",
			Help: "Количество отправленных сообщений с директориями",
		},
		[]string{"instance"},
	)
	kafkaMessagesSentFilesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_kafka_messages_sent_files_total",
			Help: "Количество отправленных сообщений с файлами",
		},
		[]string{"instance"},
	)
	kafkaMessagesReturnedDirsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_kafka_messages_returned_directories_total",
			Help: "Количество возвращенных сообщений с директориями",
		},
		[]string{"instance"},
	)
	kafkaMessagesSentScanFilesCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_kafka_messages_sent_scan_files_count_total",
			Help: "Количество отправленных сообщений с подсчетом файлов",
		},
		[]string{"instance"},
	)
	kafkaMessagesSentScanDirsCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_kafka_messages_sent_scan_dirs_count_total",
			Help: "Количество отправленных сообщений с подсчетом директорий",
		},
		[]string{"instance"},
	)
	kafkaMessagesSentCompletedDirsCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_kafka_messages_sent_completed_dirs_count_total",
			Help: "Количество отправленных сообщений с подсчетом завершенных директорий",
		},
		[]string{"instance"},
	)
	kafkaSendErrorsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "directory_lister_kafka_send_errors_total",
			Help: "Количество ошибок при отправке сообщений в Kafka",
		},
		[]string{"instance"},
	)

	// Регистрация векторов
	prometheus.MustRegister(
		receivedMessagesVec,
		readErrorsVec,
		processingDurationVec,
		ftpReconnectionsVec,
		directoriesProcessedVec,
		filesFoundVec,
		ftpListDurationVec,
		kafkaMessagesSentDirsVec,
		kafkaMessagesSentFilesVec,
		kafkaMessagesReturnedDirsVec,
		kafkaMessagesSentScanFilesCountVec,
		kafkaMessagesSentScanDirsCountVec,
		kafkaMessagesSentCompletedDirsCountVec,
		kafkaSendErrorsVec,
	)

	// Инициализация метрик с лейблом instance
	ReceivedMessages = receivedMessagesVec.WithLabelValues(instance)
	ReadErrors = readErrorsVec.WithLabelValues(instance)
	ProcessingDuration = processingDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	FtpReconnections = ftpReconnectionsVec.WithLabelValues(instance)
	DirectoriesProcessed = directoriesProcessedVec.WithLabelValues(instance)
	FilesFound = filesFoundVec.WithLabelValues(instance)
	FtpListDuration = ftpListDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	KafkaMessagesSentDirectories = kafkaMessagesSentDirsVec.WithLabelValues(instance)
	KafkaMessagesSentFiles = kafkaMessagesSentFilesVec.WithLabelValues(instance)
	KafkaMessagesReturnedDirs = kafkaMessagesReturnedDirsVec.WithLabelValues(instance)
	KafkaMessagesSentScanFilesCount = kafkaMessagesSentScanFilesCountVec.WithLabelValues(instance)
	KafkaMessagesSentScanDirsCount = kafkaMessagesSentScanDirsCountVec.WithLabelValues(instance)
	KafkaMessagesSentCompletedDirsCount = kafkaMessagesSentCompletedDirsCountVec.WithLabelValues(instance)
	KafkaSendErrors = kafkaSendErrorsVec.WithLabelValues(instance)
}
