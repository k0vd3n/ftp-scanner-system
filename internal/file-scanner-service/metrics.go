package filescannerservice

import (
	"ftp-scanner_try2/config"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Метрики для KafkaHandler file scanner
	ReceivedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "file_scanner_service_counter",
		Help: "Количество полученных сообщений для сканирования файлов",
	})
	DownloadDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "file_scanner_service_histogram",
		Help:    "Время скачивания файла с FTP-сервера",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
	FtpReconnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "file_scanner_service_counter",
		Help: "Количество событий переподключения к FTP",
	})
	ScanDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "file_scanner_service_histogram",
		Help:    "Время сканирования файла",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
	ResultMessagesSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "file_scanner_service_counter",
		Help: "Количество отправленных сообщений с результатами сканирования",
	})
	CompletedFilesSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "file_scanner_service_counter",
		Help: "Количество отправленных сообщений с количеством завершенных файлов",
	})
	ErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "file_scanner_service_counter",
		Help: "Количество ошибок при обработке файлов",
	})
)

func InitMetrics() {
	prometheus.MustRegister(
		ReceivedMessages,
		DownloadDuration,
		FtpReconnections,
		ScanDuration,
		ResultMessagesSent,
		CompletedFilesSent,
		ErrorCounter,
	)
}

func PushMetrics(cfg *config.PushGatewayConfig) {
	err := push.New(cfg.URL, cfg.JobName).
		Collector(ReceivedMessages).
		Collector(DownloadDuration).
		Collector(FtpReconnections).
		Collector(ScanDuration).
		Collector(ResultMessagesSent).
		Collector(CompletedFilesSent).
		Collector(ErrorCounter).
		Grouping("instance", cfg.Instance).
		Push()
	if err != nil {
		log.Printf("file scanner service metrics: Ошибка при отправке метрик в Pushgateway: %v", err)
	} else {
		log.Printf("file scanner service metrics: Метрики отправлены в Pushgateway")
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
