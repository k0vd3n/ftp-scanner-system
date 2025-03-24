package reportservice

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Интерфейс для хранилища отчетов
type ReportStorage interface {
	SaveReport(scanID string, data interface{}) (string, error)
}

// Локальное хранилище (файловая система)
type FileReportStorage struct {
	basePath string
}

func NewFileReportStorage(basePath string) *FileReportStorage {
	return &FileReportStorage{basePath: basePath}
}

func (s *FileReportStorage) SaveReport(scanID string, data interface{}) (string, error) {
	log.Printf("scan-reports-service storage: SaveReport: Сохранение отчета для scan_id=%s", scanID)
	filePath := fmt.Sprintf("%s/%s_timestamp_%s.json", s.basePath, scanID, fmt.Sprintf("%d", time.Now().UnixNano()))

	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", err
	}

	log.Printf("scan-reports-service storage: SaveReport: Сохранение отчета для scan_id=%s завершено", scanID)
	return fmt.Sprintf("http://example.com/reports/%s.json", scanID), nil
}
