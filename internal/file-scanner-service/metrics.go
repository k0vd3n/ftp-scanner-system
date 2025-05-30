package filescannerservice

import (
	"context"
	"encoding/binary"
	"ftp-scanner_try2/internal/mongodb"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
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

// NewPusher создает объект Pusher для отправки метрик в Pushgateway.
// Параметры:
//   - pushgatewayURL: URL адрес Pushgateway (например, "http://localhost:9091").
//   - job: Название задания (job) для группировки метрик в Pushgateway.
//   - instance: Значение instance, которое должно совпадать с переданным в InitMetrics.
//
// Возвращает настроенный объект Pusher, готовый к отправке метрик.
func NewPusher(pushgatewayURL, job, instance string) *push.Pusher {
	return push.New(pushgatewayURL, job).
		// Grouping("instance", instance).
		Gatherer(prometheus.DefaultGatherer)
}

// // GatherMetrics собирает все метрики из DefaultGatherer.
// func GatherMetrics() ([]models.Metric, error) {
// 	families, err := prometheus.DefaultGatherer.Gather()
// 	if err != nil {
// 		return nil, err
// 	}

// 	now := time.Now().UTC()
// 	var result []models.Metric

// 	for _, fam := range families {
// 		for _, m := range fam.Metric {
// 			labels := make(map[string]string, len(m.Label))
// 			for _, lp := range m.Label {
// 				labels[lp.GetName()] = lp.GetValue()
// 			}

// 			var val float64
// 			switch fam.GetType() {
// 			case dto.MetricType_COUNTER:
// 				val = m.GetCounter().GetValue()
// 			case dto.MetricType_GAUGE:
// 				val = m.GetGauge().GetValue()
// 			case dto.MetricType_HISTOGRAM:
// 				// сохраняем сумму событий
// 				val = m.GetHistogram().GetSampleSum()
// 			default:
// 				continue
// 			}

// 			result = append(result, models.Metric{
// 				Name:      fam.GetName(),
// 				Labels:    labels,
// 				Value:     val,
// 				Timestamp: now,
// 			})
// 		}
// 	}

// 	return result, nil
// }

// // GetMetricVecByName возвращает по имени метрики либо CounterVec, либо HistogramVec.
// // Если метрика не найдена, возвращает (nil, nil).
// func GetMetricVecByName(name string) (c *prometheus.CounterVec, h *prometheus.HistogramVec) {
//     switch name {
//     // === CounterVec ===
//     case "file_scanner_received_messages_total":
//         return receivedMessagesVec, nil
//     case "file_scanner_saved_files_total":
//         return savedFilesCounterVec, nil
//     case "file_scanner_ftp_reconnections_total":
//         return ftpReconnectionsVec, nil
//     case "file_scanner_result_messages_sent_total":
//         return resultMessagesSentVec, nil
//     case "file_scanner_completed_files_sent_total":
//         return sendedCompletedFilesVec, nil
//     case "file_scanner_errors_total":
//         return errorCounterVec, nil
//     case "file_scanner_returned_files_total":
//         return returnedFilesVec, nil
//     case "file_scanner_downloaded_files_total":
//         return downloadedFilesVec, nil
//     case "file_scanner_scanned_files_total":
//         return scannedFilesVec, nil
//     case "file_scanner_sended_files_total":
//         return sendedScanFilesResultsVec, nil

//     // === HistogramVec ===
//     case "file_scanner_download_duration_seconds":
//         return nil, downloadDurationVec
//     case "file_scanner_saving_duration_seconds":
//         return nil, savingDurationVec
//     case "file_scanner_scan_duration_seconds":
//         return nil, scanDurationVec
//     case "file_scanner_return_error_duration_seconds":
//         return nil, returnErrorDurationVec
//     case "file_scanner_returning_files_duration_seconds":
//         return nil, returningFilesDurationVec
//     case "file_scanner_error_sending_scan_files_result_duration_seconds":
//         return nil, errorSendingScanFilesResultDuration
//     case "file_scanner_sending_scan_files_results_duration_seconds":
//         return nil, sendingScanFilesResultsDurationVec
//     case "file_scanner_error_sending_completed_files_duration_seconds":
//         return nil, errorSendingCompletedFilesDurationVec
//     case "file_scanner_sending_completed_files_duration_seconds":
//         return nil, sendingCompletedFilesDurationVec

//     default:
//         return nil, nil
//     }
// }

// SaveMetricsToMongo собирает все метрики из DefaultGatherer, сериализует их и сохраняет в MongoDB.
func SaveMetricsToMongo(ctx context.Context, repo mongodb.MetricRepository, instance string) error {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return err
	}
	var buf []byte
	for _, mf := range mfs {
		b, err := proto.Marshal(mf)
		if err != nil {
			return err
		}
		// префикс-длина (4 байта LE)
		tmp := make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, uint32(len(b)))
		buf = append(buf, tmp...)
		buf = append(buf, b...)
	}
	return repo.Save(ctx, instance, buf)
}

// LoadMetricsFromMongo загружает сохраненные счётчики из MongoDB и добавляет их в текущие метрики.
func LoadMetricsFromMongo(ctx context.Context, repo mongodb.MetricRepository, instance string) error {
	payload, err := repo.Load(ctx, instance)
	if err != nil {
		return err
	}
	offset := 0
	for offset < len(payload) {
		length := int(binary.LittleEndian.Uint32(payload[offset:]))
		offset += 4
		mf := &dto.MetricFamily{}
		if err := proto.Unmarshal(payload[offset:offset+length], mf); err != nil {
			return err
		}
		offset += length

		switch mf.GetType() {
		case dto.MetricType_COUNTER:
			for _, m := range mf.Metric {
				for _, lp := range m.Label {
					if lp.GetName() == "instance" && lp.GetValue() == instance {
						value := m.GetCounter().GetValue()
						switch mf.GetName() {
						case "file_scanner_received_messages_total":
							ReceivedMessages.Add(value)
						case "file_scanner_saved_files_total":
							SavedFilesCounter.Add(value)
						case "file_scanner_ftp_reconnections_total":
							FtpReconnections.Add(value)
						case "file_scanner_result_messages_sent_total":
							ResultMessagesSent.Add(value)
						case "file_scanner_completed_files_sent_total":
							SendedCompletedFiles.Add(value)
						case "file_scanner_errors_total":
							ErrorCounter.Add(value)
						case "file_scanner_returned_files_total":
							ReturnedFiles.Add(value)
						case "file_scanner_downloaded_files_total":
							DownloadedFiles.Add(value)
						case "file_scanner_scanned_files_total":
							ScannedFiles.Add(value)
						case "file_scanner_sended_files_total":
							SendedScanFilesResults.Add(value)
						}
					}
				}
			}
		}
	}
	return nil
}

// // preloadedGatherer отдаёт при Gather() заранее загруженные MetricFamily.
// type preloadedGatherer struct {
// 	mfs []*dto.MetricFamily
// }

// func (g *preloadedGatherer) Gather() ([]*dto.MetricFamily, error) {
// 	return g.mfs, nil
// }
