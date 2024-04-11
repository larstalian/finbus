package rest

import (
	"encoding/json"
	"finbus/internal/services"
	"net/http"
)

type BusHandler interface {
	HandleQueryBusesNear(w http.ResponseWriter, r *http.Request)
}

type busHandler struct {
	service services.BusDataService
}

func NewBusHandler(service services.BusDataService) BusHandler {
	return &busHandler{service: service}
}

// HandleQueryBusesNear processes the API request for querying buses near specific coordinates.
func (h *busHandler) HandleQueryBusesNear(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	if lat == "" || lon == "" {
		http.Error(w, "Latitude and longitude are required", http.StatusBadRequest)
		return
	}

	buses, err := h.service.QueryBusesNear(lat, lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(buses)
}
