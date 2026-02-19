package node

import (
	"encoding/json"
	"net/http"
)

type Status struct {
	Battery  BatteryInfo  `json:"battery"`
	Location LocationInfo `json:"location"`
	Activity string       `json:"activity"`
	Wifi     WifiInfo     `json:"wifi"`
}

type BatteryInfo struct {
	Level    int  `json:"level"`
	Charging bool `json:"charging"`
}

type LocationInfo struct {
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Accuracy float64 `json:"accuracy"`
}

type WifiInfo struct {
	SSID      string `json:"ssid"`
	Connected bool   `json:"connected"`
}

type Server struct {
	status Status
	mux    *http.ServeMux
}

func New() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /status", s.handleStatus)
	s.mux.HandleFunc("POST /notify", s.handleNotify)
	s.mux.HandleFunc("GET /pair", s.handlePair)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.status)
}

func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
	// For now, just Accept
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePair(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

// UpdateStatus allows updating the current sensor state (used by companion app)
func (s *Server) UpdateStatus(st Status) {
	s.status = st
}
