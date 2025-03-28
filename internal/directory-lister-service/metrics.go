package directorylisterservice

import (
	"ftp-scanner_try2/config"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	ReceivedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "directory_lister_received_messages_total",
		Help: "Количество успешно полученных сообщений из Kafka",
	})
	ReadErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "directory_lister_read_errors_total",
		Help: "Количество ошибок при чтении сообщений из Kafka",
	})
	ProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "directory_lister_processing_duration_seconds",
		Help:    "Время обработки одного сообщения",
		Buckets: prometheus.ExponentialBuckets(0.00005, 2, 20),
	})
	FtpReconnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ftp_reconnections_total",
		Help: "Количество событий переподключения к FTP",
	})
	DirectoriesProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "directories_processed_total",
		Help: "Количество обработанных директорий",
	})
	FilesFound = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "files_found_total",
		Help: "Количество найденных файлов",
	})
	FtpListDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ftp_list_directory_duration_seconds",
		Help:    "Время листинга директории",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
	KafkaMessagesSentDirectories = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_sent_directories_total",
		Help: "Количество отправленных сообщений с директориями",
	})
	KafkaMessagesSentFiles = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_sent_files_total",
		Help: "Количество отправленных сообщений с файлами",
	})
	KafkaMessagesSentCounts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_sent_counts_total",
		Help: "Количество отправленных сообщений с подсчетом элементов",
	})
	KafkaSendErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kafka_send_errors_total",
		Help: "Количество ошибок при отправке сообщений в Kafka",
	})
)

// InitMetrics регистрирует метрики для сервиса directory-lister-service
func InitMetrics() {
	prometheus.MustRegister(
		ReceivedMessages,
		ReadErrors,
		ProcessingDuration,
		FtpReconnections,
		DirectoriesProcessed,
		FilesFound,
		FtpListDuration,
		KafkaMessagesSentDirectories,
		KafkaMessagesSentFiles,
		KafkaMessagesSentCounts,
		KafkaSendErrors,
	)
}

// PushMetrics отправляет метрики в PushGateway
// Запуск этой функции происходит в отдельной горутине
func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(ReceivedMessages).
		Collector(ReadErrors).
		Collector(ProcessingDuration).
		Collector(FtpReconnections).
		Collector(DirectoriesProcessed).
		Collector(FilesFound).
		Collector(FtpListDuration).
		Collector(KafkaMessagesSentDirectories).
		Collector(KafkaMessagesSentFiles).
		Collector(KafkaMessagesSentCounts).
		Collector(KafkaSendErrors).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("Ошибка при отправке метрик в Pushgateway: %v", err)
	} else {
		log.Printf("Метрики отправлены в Pushgateway")
	}
}

// StartPushLoop запускает периодическую отправку метрик в Pushgateway.
func StartPushLoop(cfg *config.PushGatewayConfig) {
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.PushInterval) * time.Second)
		for range ticker.C {
			PushMetrics(cfg)
		}
	}()
}
