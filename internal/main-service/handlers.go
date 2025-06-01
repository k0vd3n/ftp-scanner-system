package mainservice

import (
	"encoding/json"
	"ftp-scanner_try2/internal/models"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (s *MainServer) HandleStartScan(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	RequestsTotal.Inc()
	var req models.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Main-service: handlers: startScan: Ошибка декодирования запроса", zap.Error(err))
		http.Error(w, "Не удалось декодировать запрос", http.StatusBadRequest)
		return
	}

	// Валидация запроса
	if req.DirectoryPath == "" {
		s.logger.Error("Main-service: handlers: startScan: directory_path не указан")
		http.Error(w, "Необходимо указать directory_path", http.StatusBadRequest)
		return
	}

	if len(req.ScanTypes) == 0 {
		s.logger.Error("Main-service: handlers: startScan: scan_types не указаны")
		http.Error(w, "Необходимо указать scan_types", http.StatusBadRequest)
		return
	}

	if req.FTPServer == "" {
		s.logger.Error("Main-service: handlers: startScan: ftp_server не указан")
		http.Error(w, "Необходимо указать ftp_server", http.StatusBadRequest)
		return
	}

	if req.FTPPort == 0 {
		s.logger.Error("Main-service: handlers: startScan: ftp_port не указан")
		http.Error(w, "Необходимо указать ftp_port", http.StatusBadRequest)
		return
	}

	if req.FTPUsername == "" {
		s.logger.Error("Main-service: handlers: startScan: ftp_username не указан")
		http.Error(w, "Необходимо указать ftp_username", http.StatusBadRequest)
		return
	}

	// Запуск сканирования
	startScanTime := time.Now()
	resp, err := s.StartScan(r.Context(), req)
	ProcessingStartScanMethod.Observe(time.Since(startScanTime).Seconds())
	if err != nil {
		s.logger.Error("Main-service: handlers: startScan: Ошибка запуска сканирования", zap.Error(err))
		http.Error(w, "Не удалось запустить сканирование", http.StatusInternalServerError)
		return
	}
	s.logger.Info("Main-service: handlers: startScan: Сканирование запущено", zap.Any("response", resp))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	duration := time.Since(start).Seconds()
	s.logger.Info("Main-service: handlers: startScan: Время запуска сканирования", zap.Float64("duration", duration))
	ProcessingScanRequest.Observe(duration)
	s.logger.Info("Main-service: handlers: startScan: Сканирование запущено", zap.Any("response", resp))
}

func (s *MainServer) HandleGetScanStatus(w http.ResponseWriter, r *http.Request) {
	startTimeRequest := time.Now()
	RequestsTotal.Inc()
	s.logger.Info("Main-service: handlers: getScanStatus: Получение статуса сканирования...")
	scanID := r.URL.Query().Get("scan_id")
	if scanID == "" {
		s.logger.Info("Main-service: handlers: getScanStatus: Необходимо указать scan_id")
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	startGetScanStatusTime := time.Now()
	// Получение статуса сканирования
	resp, err := s.GetScanStatus(r.Context(), scanID)
	GRPCStatusCallDuration.Observe(time.Since(startGetScanStatusTime).Seconds())
	if err != nil {
		s.logger.Error("Main-service: handlers: getScanStatus: Ошибка получения статуса сканирования", zap.Error(err))
		http.Error(w, "Ошибка получения статуса сканирования", http.StatusInternalServerError)
		return
	}
	s.logger.Info("Main-service: handlers: getScanStatus: Статус сканирования получен", zap.Any("response", resp))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	ProcessingStatusRequest.Observe(time.Since(startTimeRequest).Seconds())
	s.logger.Info("Main-service: handlers: getScanStatus: Статус сканирования получен", zap.Any("response", resp))
}

func (s *MainServer) HandleGenerateReport(w http.ResponseWriter, r *http.Request) {
	startTimeRequest := time.Now()
	RequestsTotal.Inc()
	s.logger.Info("Main-service: handlers: getReport: Получение отчета...")
	scanID := r.URL.Query().Get("scan_id")
	s.logger.Info("Main-service: handlers: getReport: получен scan_id", zap.String("scan_id", scanID))
	if scanID == "" {
		s.logger.Info("Main-service: handlers: getReport: Необходимо указать scan_id")
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	// Получение отчета
	startGenerateReportTime := time.Now()
	resp, err := s.GenerateReport(r.Context(), scanID)
	GRPCReportCallDuration.Observe(time.Since(startGenerateReportTime).Seconds())
	if err != nil {
		s.logger.Error("Main-service: handlers: getReport: Ошибка генерации отчета", zap.Error(err))
		http.Error(w, "Ошибка генерации отчета", http.StatusInternalServerError)
		return
	}
	s.logger.Info("Main-service: handlers: getReport: Отчет успешно сгенерирован", zap.Any("response", resp))

	w.Header().Set("Content-Type", "application/json")
	resp.Message = "Отчет успешно сгенерирован"
	json.NewEncoder(w).Encode(resp)
	ProcessingReportRequest.Observe(time.Since(startTimeRequest).Seconds())
	s.logger.Info("Main-service: handlers: getReport: Отчет успешно сгенерирован", zap.Any("response", resp))
}

func (s *MainServer) ApiHandleGetDirectory(w http.ResponseWriter, r *http.Request) {
	startTimeRequest := time.Now()
	RequestsTotal.Inc()
	s.logger.Info("Main-service: handlers: getReport: Получение отчета...")
	scanID := r.URL.Query().Get("scan_id")
	s.logger.Info("Main-service: handlers: getReport: получен scan_id", zap.String("scan_id", scanID))
	if scanID == "" {
		s.logger.Info("Main-service: handlers: getReport: Необходимо указать scan_id")
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	dirPath := r.URL.Query().Get("Path")
	s.logger.Info("Main-service: handlers: getReport: получен Path", zap.String("Path", dirPath))

	// Получение отчета
	startGetReportTime := time.Now()
	resp, err := s.GetReport(r.Context(), scanID, dirPath)
	GRPCReportCallDuration.Observe(time.Since(startGetReportTime).Seconds())
	if err != nil {
		s.logger.Error("Main-service: handlers: getReport: Ошибка генерации отчета", zap.Error(err))
		http.Error(w, "Ошибка генерации отчета", http.StatusInternalServerError)
		return
	}
	s.logger.Info("Main-service: handlers: getReport: Отчет успешно сгенерирован", zap.Any("response", resp))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("Main-service: handlers: getReport: Ошибка отправки отчета", zap.Error(err))
		http.Error(w, "Ошибка отправки отчета", http.StatusInternalServerError)
		return
	}
	ProcessingReportRequest.Observe(time.Since(startTimeRequest).Seconds())
	s.logger.Info("Main-service: handlers: getReport: Отчет успешно сгенерирован", zap.Any("response", resp))
}

func (s *MainServer) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	startTimeRequest := time.Now()
	RequestsTotal.Inc()
	s.logger.Info("Main-service: handlers: getReport: Получение отчета...")
	scanID := r.URL.Query().Get("scan_id")
	s.logger.Info("Main-service: handlers: getReport: получен scan_id", zap.String("scan_id", scanID))
	if scanID == "" {
		s.logger.Info("Main-service: handlers: getReport: Необходимо указать scan_id")
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	// Здесь надо вывести index.html
	http.ServeFile(w, r, s.config.HTTP.WebPath)

	ProcessingReportRequest.Observe(time.Since(startTimeRequest).Seconds())
	s.logger.Info("Main-service: handlers: getReport: Отчет успешно сгенерирован")
}
