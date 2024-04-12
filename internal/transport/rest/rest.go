package rest

import (
	"encoding/json"
	"finbus/internal/models"
	"finbus/internal/services"
	"net/http"
	"strconv"
)

type BusHandler interface {
	HandleQueryBusesNear(w http.ResponseWriter, r *http.Request)
	HandleGetBusesFromStops(writer http.ResponseWriter, request *http.Request)
}
type busHandler struct {
	service services.BusDataService
}

// NewBusHandler creates a new BusHandler
func NewBusHandler(service services.BusDataService) BusHandler {
	return &busHandler{service: service}
}

// HandleQueryBusesNear processes the API request for querying buses near specific coordinates.
func (h *busHandler) HandleQueryBusesNear(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	if latStr == "" || lonStr == "" {
		http.Error(w, "Latitude and longitude are required", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude value", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude value", http.StatusBadRequest)
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

// HandleGetBusesFromStops processes the API request for querying buses from specific stops.
func (h *busHandler) HandleGetBusesFromStops(w http.ResponseWriter, r *http.Request) {
	var stopsData []models.BusData

	err := json.NewDecoder(r.Body).Decode(&stopsData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(stopsData) == 0 {
		http.Error(w, "No stops provided", http.StatusBadRequest)
		return
	}

	busData, err := h.service.GetBusQueryFromStops(stopsData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(busData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

var _ BusHandler = (*busHandler)(nil)
