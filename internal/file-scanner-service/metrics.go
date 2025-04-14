package filescannerservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	receivedMessagesVec                   *prometheus.CounterVec
	ReceivedMessages                      prometheus.Counter
	downloadDurationVec                   *prometheus.HistogramVec
	DownloadDuration                      prometheus.Histogram
	savingDurationVec                     *prometheus.HistogramVec
	SavingDuration                        prometheus.Histogram
	savedFilesCounterVec                  *prometheus.CounterVec
	SavedFilesCounter                     prometheus.Counter
	ftpReconnectionsVec                   *prometheus.CounterVec
	FtpReconnections                      prometheus.Counter
	scanDurationVec                       *prometheus.HistogramVec
	ScanDuration                          prometheus.Histogram
	resultMessagesSentVec                 *prometheus.CounterVec
	ResultMessagesSent                    prometheus.Counter
	sendedCompletedFilesVec               *prometheus.CounterVec
	SendedCompletedFiles                  prometheus.Counter
	errorCounterVec                       *prometheus.CounterVec
	ErrorCounter                          prometheus.Counter
	returnErrorDurationVec                *prometheus.HistogramVec
	ReturnErrorDuration                   prometheus.Histogram
	returnedFilesVec                      *prometheus.CounterVec
	ReturnedFiles                         prometheus.Counter
	returningFilesDurationVec             *prometheus.HistogramVec
	ReturningFilesDuration                prometheus.Histogram
	downloadedFilesVec                    *prometheus.CounterVec
	DownloadedFiles                       prometheus.Counter
	scannedFilesVec                       *prometheus.CounterVec
	ScannedFiles                          prometheus.Counter
	sendedScanFilesResultsVec             *prometheus.CounterVec
	SendedScanFilesResults                prometheus.Counter
	errorSendingScanFilesResultDuration   *prometheus.HistogramVec
	ErrorSendingScanFilesResultDuration   prometheus.Histogram
	sendingScanFilesResultsDurationVec    *prometheus.HistogramVec
	SendingScanFilesResultsDuration       prometheus.Histogram
	errorSendingCompletedFilesDurationVec *prometheus.HistogramVec
	ErrorSendingCompletedFilesDuration    prometheus.Histogram
	sendingCompletedFilesDurationVec      *prometheus.HistogramVec
	SendingCompletedFilesDuration         prometheus.Histogram
)

// InitMetrics регистрирует метрики с лейблом instance
func InitMetrics(instance string) {
	// Инициализация векторов метрик
	receivedMessagesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_received_messages_total",
			Help: "Количество успешно полученных сообщений из Kafka",
		},
		[]string{"instance"},
	)

	downloadDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_download_duration_seconds",
			Help:    "Время скачивания файлов",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	savingDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_saving_duration_seconds",
			Help:    "Время сохранения файлов",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	savedFilesCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_saved_files_total",
			Help: "Количество сохраненных файлов",
		},
		[]string{"instance"},
	)

	ftpReconnectionsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_ftp_reconnections_total",
			Help: "Количество переподключений к FTP",
		},
		[]string{"instance"},
	)

	scanDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_scan_duration_seconds",
			Help:    "Время сканирования файлов",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"instance"},
	)

	resultMessagesSentVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_result_messages_sent_total",
			Help: "Количество отправленных сообщений с результатами",
		},
		[]string{"instance"},
	)

	sendedCompletedFilesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_completed_files_sent_total",
			Help: "Количество обработанных файлов",
		},
		[]string{"instance"},
	)

	errorCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_errors_total",
			Help: "Количество ошибок при обработке файлов",
		},
		[]string{"instance"},
	)

	returnErrorDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_return_error_duration_seconds",
			Help:    "Время возвращения ошибок",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	returnedFilesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_returned_files_total",
			Help: "Количество возвращенных файлов",
		},
		[]string{"instance"},
	)

	returningFilesDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_returning_files_duration_seconds",
			Help:    "Время возвращения файлов",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	downloadedFilesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_downloaded_files_total",
			Help: "Количество загруженных файлов",
		},
		[]string{"instance"},
	)

	scannedFilesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_scanned_files_total",
			Help: "Количество сканированных файлов",
		},
		[]string{"instance"},
	)

	sendedScanFilesResultsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_scanner_sended_files_total",
			Help: "Количество отправленных файлов",
		},
		[]string{"instance"},
	)

	errorSendingScanFilesResultDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_error_sending_scan_files_result_duration_seconds",
			Help:    "Время отправки результатов сканирования",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	sendingScanFilesResultsDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_sending_scan_files_results_duration_seconds",
			Help:    "Время отправки результатов сканирования",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	errorSendingCompletedFilesDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_error_sending_completed_files_duration_seconds",
			Help:    "Время отправки результатов сканирования",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	sendingCompletedFilesDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_scanner_sending_completed_files_duration_seconds",
			Help:    "Время отправки результатов сканирования",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"instance"},
	)

	// Регистрация векторов
	prometheus.MustRegister(
		receivedMessagesVec,
		downloadDurationVec,
		savedFilesCounterVec,
		savingDurationVec,
		ftpReconnectionsVec,
		scanDurationVec,
		resultMessagesSentVec,
		sendedCompletedFilesVec,
		errorCounterVec,
		returningFilesDurationVec,
		returnErrorDurationVec,
		returnedFilesVec,
		downloadedFilesVec,
		scannedFilesVec,
		sendedScanFilesResultsVec,
		errorSendingScanFilesResultDuration,
		sendingScanFilesResultsDurationVec,
		errorSendingCompletedFilesDurationVec,
		sendingCompletedFilesDurationVec,
	)

	// Инициализация метрик с лейблом instance
	ReceivedMessages = receivedMessagesVec.WithLabelValues(instance)
	DownloadDuration = downloadDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	SavingDuration = savingDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	SavedFilesCounter = savedFilesCounterVec.WithLabelValues(instance)
	FtpReconnections = ftpReconnectionsVec.WithLabelValues(instance)
	ScanDuration = scanDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ResultMessagesSent = resultMessagesSentVec.WithLabelValues(instance)
	SendedCompletedFiles = sendedCompletedFilesVec.WithLabelValues(instance)
	ErrorCounter = errorCounterVec.WithLabelValues(instance)
	ReturnErrorDuration = returnErrorDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ReturnedFiles = returnedFilesVec.WithLabelValues(instance)
	ReturningFilesDuration = returningFilesDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	DownloadedFiles = downloadedFilesVec.WithLabelValues(instance)
	ScannedFiles = scannedFilesVec.WithLabelValues(instance)
	SendedScanFilesResults = sendedScanFilesResultsVec.WithLabelValues(instance)
	ErrorSendingScanFilesResultDuration = errorSendingScanFilesResultDuration.WithLabelValues(instance).(prometheus.Histogram)
	SendingScanFilesResultsDuration = sendingScanFilesResultsDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	ErrorSendingCompletedFilesDuration = errorSendingCompletedFilesDurationVec.WithLabelValues(instance).(prometheus.Histogram)
	SendingCompletedFilesDuration = sendingCompletedFilesDurationVec.WithLabelValues(instance).(prometheus.Histogram)
}
