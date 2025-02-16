package models

// DirectoryScanMessage представляет собой сообщение для топика `directories-to-scan`.
type DirectoryScanMessage struct {
	ScanID      string   `json:"scan_id"`      // Идентификатор сканирования
	DirectoryPath string `json:"directory_path"` // Путь к директории
	ScanTypes    []string `json:"scan_types"`   // Типы сканирования
}

// FileScanMessage представляет собой сообщение для топика `files-to-scan`.
type FileScanMessage struct {
	ScanID   string `json:"scan_id"`   // Идентификатор сканирования
	FilePath string `json:"file_path"` // Путь к файлу
	ScanType string `json:"scan_type"` // Тип сканирования
}

// ScanResultMessage представляет собой сообщение для топика `scan-results`.
type ScanResultMessage struct {
	ScanID   string `json:"scan_id"`   // Идентификатор сканирования
	FilePath string `json:"file_path"` // Путь к файлу
	ScanType string `json:"scan_type"` // Тип сканирования
	Result   string `json:"result"`    // Результат сканирования
}

// CountMessage представляет собой сообщение для топиков счетчиков.
type CountMessage struct {
	ScanID string `json:"scan_id"` // Идентификатор сканирования
	Number int    `json:"number"`  // Число для суммирования
}