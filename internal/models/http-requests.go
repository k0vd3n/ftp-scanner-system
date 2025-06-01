package models

// ScanRequest представляет собой запрос на сканирование.
type ScanRequest struct {
	DirectoryPath string   `json:"directory_path"` // Путь к директории
	ScanTypes     []string `json:"scan_types"`     // Типы сканирования
	FTPServer     string   `json:"ftp_server"`     // Адрес FTP-сервера
	FTPPort       int      `json:"ftp_port"`       // Порт FTP-сервера
	FTPUsername   string   `json:"ftp_username"`   // Имя пользователя FTP
	FTPPassword   string   `json:"ftp_password"`   // Пароль FTP
}

// ScanResponsоe представляет сбой ответ на запрос сканирования.
type ScanResponse struct {
	ScanID  string `json:"scan_id"` // Идентификатор сканирования
	Status  string `json:"status"`  // Статус сканирования
	Message string `json:"message"` // Сообщение
	// Ниже под вопросом
	// FTPConnection FTPConnection `json:"ftp_connection"` // Информация о подключении к FTP
	StartTime string `json:"start_time"` // Время начала сканирования
}

// FTPConnection представляет собой информацию о подключении к FTP.
type FTPConnection struct {
	Server   string `json:"ftp_server"`   // Адрес FTP-сервера
	Port     int    `json:"ftp_port"`     // Порт FTP-сервера
	Username string `json:"ftp_username"` // Имя пользователя FTP
	Password string `json:"ftp_password"` // Пароль FTP
}

// StatusRequest представляет собой запрос на получение состояния сканирования.
type StatusRequest struct {
	ScanID string `json:"scan_id"` // Идентификатор сканирования
}

// StatusResponse представляет собой ответ с состоянием сканирования.
type StatusResponse struct {
	ScanID             string `json:"scan_id"`             // Идентификатор сканирования
	Status             string `json:"status"`              // Статус сканирования
	DirectoriesScanned int    `json:"directories_scanned"` // Количество обработанных директорий
	FilesScanned       int    `json:"files_scanned"`       // Количество обработанных файлов
	DirectoriesFound   int    `json:"directories_found"`   // Количество найденных директорий
	FilesFound         int    `json:"files_found"`         // Количество найденных файлов
	StartTime          string `json:"start_time"`          // Время начала сканирования
}

// ReportRequest представляет собой запрос на получение отчета.
type ReportRequest struct {
	ScanID string `json:"scan_id"` // Идентификатор сканирования
}

// ReportResponse представляет собой ответ с ссылкой на отчет.
type ReportResponse struct {
	ScanID  string `json:"scan_id"` // Идентификатор сканирования
	Message string `json:"message"` // Сообщение
}

// Объект-ответ для динамического запроса,
// включающий ScanID и найденную директорию.
type DirectoryResponse struct {
	ScanID    string    `json:"ScanID"`
	Directory Directory `json:"Directory"`
}