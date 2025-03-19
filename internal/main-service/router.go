package mainservice

import (
	"net/http"
)

func (s *MainServer) SetupRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/scan/start", s.HandleStartScan)
	router.HandleFunc("/scan/status", s.HandleGetScanStatus)
	router.HandleFunc("/scan/report", s.HandleGetReport)

	return router
}
