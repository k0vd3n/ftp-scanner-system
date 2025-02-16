package models

import (
	"time"
)

// ScanReport представляет собой структуру для хранения отчета о сканировании.
type ScanReport struct {
	ScanID      string      `bson:"scan_id"`      // Идентификатор сканирования
	Directories []Directory `bson:"directories"`  // Иерархия директорий и файлов
	CreatedAt   time.Time   `bson:"created_at"`   // Время создания отчета
}

// Directory представляет собой директорию в иерархии.
type Directory struct {
	Directory    string      `bson:"directory"`     // Название директории
	Subdirectory []Directory `bson:"subdirectory"`  // Вложенные директории
	Files        []File      `bson:"files"`         // Файлы в директории
}

// File представляет собой файл в директории.
type File struct {
	Path        string       `bson:"path"`         // Полный путь к файлу
	ScanResults []ScanResult `bson:"scan_results"` // Результаты сканирования
}

// ScanResult представляет собой результат сканирования файла.
type ScanResult struct {
	Type   string `bson:"type"`   // Тип сканирования
	Result string `bson:"result"` // Результат сканирования
}