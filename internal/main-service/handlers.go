package mainservice

import (
	"encoding/json"
	"ftp-scanner_try2/internal/models"
	"net/http"
)

func (s *MainServer) HandleStartScan(w http.ResponseWriter, r *http.Request) {
	var req models.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация запроса
	if req.DirectoryPath == "" {
		http.Error(w, "directory_path is required", http.StatusBadRequest)
		return
	}

	if len(req.ScanTypes) == 0 {
		http.Error(w, "scan_types is required", http.StatusBadRequest)
		return
	}

	if req.FTPServer == "" {
		http.Error(w, "ftp_server is required", http.StatusBadRequest)
		return
	}

	if req.FTPPort == 0 {
		http.Error(w, "ftp_port is required", http.StatusBadRequest)
		return
	}

	if req.FTPUsername == "" {
		http.Error(w, "ftp_username is required", http.StatusBadRequest)
		return
	}

	// Запуск сканирования
	resp, err := s.StartScan(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *MainServer) HandleGetScanStatus(w http.ResponseWriter, r *http.Request) {
	scanID := r.URL.Query().Get("scan_id")
	if scanID == "" {
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	// Получение статуса сканирования
	resp, err := s.GetScanStatus(r.Context(), scanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *MainServer) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	scanID := r.URL.Query().Get("scan_id")
	if scanID == "" {
		http.Error(w, "scan_id is required", http.StatusBadRequest)
		return
	}

	// Получение отчета
	resp, err := s.GetReport(r.Context(), scanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}