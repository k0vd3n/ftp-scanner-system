package mainservice

import (
	"encoding/json"
	"ftp-scanner_try2/internal/models"
	"log"
	"net/http"
	"time"
)

func (s *MainServer) HandleStartScan(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	RequestsTotal.Inc()
	var req models.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Main-service: handlers: startScan: Ошибка декодирования запроса: %v", err)
		http.Error(w, "Не удалось декодировать запрос", http.StatusBadRequest)
		return
	}

	// Валидация запроса
	if req.DirectoryPath == "" {
		log.Printf("Main-service: handlers: startScan: directory_path не указан")
		http.Error(w, "Необходимо указать directory_path", http.StatusBadRequest)
		return
	}

	if len(req.ScanTypes) == 0 {
		log.Printf("Main-service: handlers: startScan: scan_types не указаны")
		http.Error(w, "Необходимо указать scan_types", http.StatusBadRequest)
		return
	}

	if req.FTPServer == "" {
		log.Printf("Main-service: handlers: startScan: ftp_server не указан")
		http.Error(w, "Необходимо указать ftp_server", http.StatusBadRequest)
		return
	}

	if req.FTPPort == 0 {
		log.Printf("Main-service: handlers: startScan: ftp_port не указан")
		http.Error(w, "Необходимо указать ftp_port", http.StatusBadRequest)
		return
	}

	if req.FTPUsername == "" {
		log.Printf("Main-service: handlers: startScan: ftp_username не указан")
		http.Error(w, "Необходимо указать ftp_username", http.StatusBadRequest)
		return
	}

	// Запуск сканирования
	startScanTime := time.Now()
	resp, err := s.StartScan(r.Context(), req)
	ProcessingStartScanMethod.Observe(time.Since(startScanTime).Seconds())
	if err != nil {
		log.Printf("Main-service: handlers: startScan: Ошибка запуска сканирования: %v", err)
		http.Error(w, "Не удалось запустить сканирование", http.StatusInternalServerError)
		return
	}
	log.Printf("Main-service: handlers: startScan: Сканирование запущено: %+v", resp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	duration := time.Since(start).Seconds()
	log.Printf("Main-service: handlers: время запуска сканирования: %f", duration)
	ProcessingScanRequest.Observe(duration)
	log.Printf("Main-service: handlers: Сканирование запущено")
}

func (s *MainServer) HandleGetScanStatus(w http.ResponseWriter, r *http.Request) {
	startTimeRequest := time.Now()
	RequestsTotal.Inc()
	log.Printf("Main-service: handlers: Получение статуса сканирования...")
	scanID := r.URL.Query().Get("scan_id")
	if scanID == "" {
		log.Printf("Main-service: handlers: getScanStatus: Необходимо указать scan_id")
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	startGetScanStatusTime := time.Now()
	// Получение статуса сканирования
	resp, err := s.GetScanStatus(r.Context(), scanID)
	GRPCStatusCallDuration.Observe(time.Since(startGetScanStatusTime).Seconds())
	if err != nil {
		log.Printf("Main-service: handlers: getScanStatus: Ошибка получения статуса сканирования: %v", err)
		http.Error(w, "Ошибка получения статуса сканирования", http.StatusInternalServerError)
		return
	}
	log.Printf("Main-service: handlers: getScanStatus: Статус сканирования получен: %+v", resp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	ProcessingStatusRequest.Observe(time.Since(startTimeRequest).Seconds())
	log.Printf("Main-service: handlers: getScanStatus: Статус сканирования отправлен")
}

func (s *MainServer) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	startTimeRequest := time.Now()
	RequestsTotal.Inc()
	log.Printf("Main-service: handlers: getReport: Получение отчета...")
	scanID := r.URL.Query().Get("scan_id")
	log.Printf("Main-service: handlers: getReport: scan_id: %s", scanID)
	if scanID == "" {
		log.Printf("Main-service: handlers: getReport: Необходимо указать scan_id")
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	// Получение отчета
	startGetReportTime := time.Now()
	resp, err := s.GetReport(r.Context(), scanID)
	GRPCReportCallDuration.Observe(time.Since(startGetReportTime).Seconds())
	if err != nil {
		log.Printf("Main-service: handlers: getReport: Ошибка получения отчета: %v", err)
		http.Error(w, "Ошибка получения отчета", http.StatusInternalServerError)
		return
	}
	log.Printf("Main-service: handlers: getReport: Отчет получен: %+v", resp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	ProcessingReportRequest.Observe(time.Since(startTimeRequest).Seconds())
	log.Printf("Main-service: handlers: getReport: Отчет отправлен")
}
