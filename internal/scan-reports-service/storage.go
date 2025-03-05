package reportservice

import (
	"encoding/json"
	"fmt"
	"os"
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
	filePath := fmt.Sprintf("%s/%s.json", s.basePath, scanID)

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

	return fmt.Sprintf("http://example.com/reports/%s.json", scanID), nil
}
