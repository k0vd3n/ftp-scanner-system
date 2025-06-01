package mainservice

import (
	"net/http"
)

func (s *MainServer) SetupRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/scan/start", s.HandleStartScan)
	router.HandleFunc("/scan/status", s.HandleGetScanStatus)
	router.HandleFunc("/scan/generate-report", s.HandleGenerateReport)
	router.HandleFunc("/scan/get-report", s.HandleGetReport)
	router.HandleFunc("/api/report", s.ApiHandleGetDirectory)

	return router
}
