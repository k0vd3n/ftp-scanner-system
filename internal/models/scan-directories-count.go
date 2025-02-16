package models

// ScanCount представляет собой структуру для хранения количества элементов.
type ElementsCount struct {
	ScanID string `bson:"scan_id"` // Идентификатор сканирования
	Count  int    `bson:"count"`   // Количество элементов
}