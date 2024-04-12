package ws

import (
	"encoding/json"
	"finbus/internal/models"
	"finbus/internal/services"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type WebSocketHandler interface {
	HandleBusUpdatesWS(w http.ResponseWriter, r *http.Request)
}

type webSocketHandler struct {
	upgrade websocket.Upgrader
	service services.BusDataService
}

// HandleBusUpdatesWS handles WebSocket connections for bus updates
func (h *webSocketHandler) HandleBusUpdatesWS(w http.ResponseWriter, r *http.Request) {
	ws, err := h.upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Close the WebSocket connection when the function returns
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}(ws)

	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Printf("Error reading initial message: %v", err)
		return
	}

	var coords models.ClientCoords
	if err := json.Unmarshal(message, &coords); err != nil {
		log.Printf("Error unmarshalling initial coordinates: %v", err)
		return
	}

	dataChannel, err := h.service.SubscribeToBusUpdates(coords)
	if err != nil {
		log.Printf("Error in subscription service: %v", err)
		return
	}

	for busData := range dataChannel {
		if err := ws.WriteJSON(busData); err != nil {
			log.Printf("Error sending data over WebSocket: %v", err)
			break
		}
	}
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(service services.BusDataService) WebSocketHandler {
	upgrade := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &webSocketHandler{upgrade: upgrade, service: service}
}

var _ WebSocketHandler = (*webSocketHandler)(nil)
