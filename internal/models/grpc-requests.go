package models

// ReportRequestGRPC представляет собой GRPC-запрос на формирование отчета.
type ReportRequestGRPC struct {
	ScanID string `json:"scan_id"` // Идентификатор сканирования
}

// ReportResponseGRPC представляет собой GRPC-ответ с ссылкой на отчет.
type ReportResponseGRPC struct {
	ScanID    string `json:"scan_id"`    // Идентификатор сканирования
	ReportURL string `json:"report_url"` // Ссылка на отчет
}

// CounterRequestGRPC представляет собой GRPC-запрос на получение значений счетчиков.
type CounterRequestGRPC struct {
	ScanID string `json:"scan_id"` // Идентификатор сканирования
}

// CounterResponseGRPC представляет собой GRPC-ответ с значениями счетчиков.
type CounterResponseGRPC struct {
	ScanID               string `json:"scan_id"`                // Идентификатор сканирования
	DirectoriesCount     int64    `json:"directories_count"`      // Количество директорий
	FilesCount           int64    `json:"files_count"`            // Количество файлов
	CompletedDirectories int64    `json:"completed_directories"`  // Количество завершенных директорий
	CompletedFiles       int64    `json:"completed_files"`        // Количество завершенных файлов
}