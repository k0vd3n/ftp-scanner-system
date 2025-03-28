package reportservice

import (
	"context"
	"ftp-scanner_try2/internal/models"
	"ftp-scanner_try2/internal/mongodb"
	"log"
	"time"
)

type ReportServiceInterface interface {
	GenerateReport(ctx context.Context, scanID string) (string, error)
}

type ReportService struct {
	repo    mongodb.ReportRepository
	storage ReportStorage
}

func NewReportService(repo mongodb.ReportRepository, storage ReportStorage) *ReportService {
	return &ReportService{repo: repo, storage: storage}
}

func (s *ReportService) GenerateReport(ctx context.Context, scanID string) (string, error) {
	log.Printf("scan-reports-service service: getReport: Получение отчета для scan_id=%s", scanID)
	reports, err := s.repo.GetReportsByScanID(ctx, scanID)
	if err != nil {
		return "", err
	}

	log.Printf("scan-reports-service service: getReport: Группировка результатов сканирования в отчет для scan_id=%s", scanID)
	groupedData := groupScanResults(reports)

	log.Printf("scan-reports-service service: getReport: Сохранение отчета для scan_id=%s", scanID)
	storageStart := time.Now()
	reportUrl, err := s.storage.SaveReport(scanID, groupedData)
	storageDuration := float64(time.Since(storageStart).Milliseconds())
	ReportStorageDuration.Observe(storageDuration)
	if err != nil {
		ReportErrorsTotal.Inc()
		return "", err
	}
	
	return reportUrl, nil
}

// Функция группировки сообщений
// Объединение нескольких JSON-деревьев в одно
func groupScanResults(reports []models.ScanReport) []models.ScanReport {
	log.Printf("scan-reports-service service: groupScanResults: Группировка результатов сканирования")
	scanMap := make(map[string]*models.ScanReport)

	for _, report := range reports {
		if _, exists := scanMap[report.ScanID]; !exists {
			scanMap[report.ScanID] = &models.ScanReport{
				ScanID:      report.ScanID,
				Directories: []models.Directory{},
			}
		}

		currentReport := scanMap[report.ScanID]
		// Для каждой директории из нового отчёта объединяем её с уже существующими
		for _, dir := range report.Directories {
			currentReport.Directories = mergeDirectorySlice(currentReport.Directories, dir)
		}
	}

	// Преобразуем map в слайс для JSON-ответа
	var groupedResults []models.ScanReport
	for _, report := range scanMap {
		groupedResults = append(groupedResults, *report)
	}
	return groupedResults
}

// mergeDirectorySlice объединяет одну директорию (src) в срез директорий (dst) по абсолютному пути.
func mergeDirectorySlice(dst []models.Directory, src models.Directory) []models.Directory {
	log.Printf("scan-reports-service service: mergeDirectorySlice: Объединение директорий")
	for i, d := range dst {
		if d.Directory == src.Directory {
			// Слияние файлов: если файл уже есть – объединяем scan_results, иначе добавляем файл.
			dst[i].Files = mergeFiles(dst[i].Files, src.Files)
			// Рекурсивно объединяем поддиректории
			for _, sub := range src.Subdirectory {
				dst[i].Subdirectory = mergeDirectorySlice(dst[i].Subdirectory, sub)
			}
			return dst
		}
	}
	// Если директории с таким путём нет – просто добавляем
	return append(dst, src)
}

// mergeFiles объединяет два среза файлов по абсолютному пути файла.
func mergeFiles(dst []models.File, src []models.File) []models.File {
	log.Printf("scan-reports-service service: mergeFiles: Объединение файлов")
	for _, fileSrc := range src {
		found := false
		for i, fileDst := range dst {
			if fileDst.Path == fileSrc.Path {
				// Объединяем результаты сканирования
				dst[i].ScanResults = append(dst[i].ScanResults, fileSrc.ScanResults...)
				found = true
				break
			}
		}
		if !found {
			dst = append(dst, fileSrc)
		}
	}
	return dst
}
