package tests

import (
	"bytes"
	"encoding/json"
	"finbus/internal/database/influxdb"
	"finbus/internal/models"
	"finbus/internal/services"
	"finbus/internal/transport/mqtt"
	"finbus/internal/transport/rest"
	"finbus/internal/transport/ws"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebSocketBusUpdates(t *testing.T) {
	busChannel := make(chan models.BusData)
	router := mux.NewRouter()

	dbManager, err := influxdb.NewBusDataManager()
	if err != nil {
		t.Errorf("Error connecting to InfluxDB, start the docker container to run this test: %v", err)
	}
	busDataSubscriber, _ := mqtt.NewBusDataSubscriber("mqtts://mqtt.digitransit.fi:8883", busChannel)
	busDataService := services.NewBusDataService(dbManager, busChannel, busDataSubscriber)
	webSocketHandler := ws.NewWebSocketHandler(busDataService)
	router.HandleFunc("/ws/bus-updates", webSocketHandler.HandleBusUpdatesWS)
	server := httptest.NewServer(router)
	defer server.Close()

	u := "ws" + server.URL[4:] + "/ws/bus-updates"

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Dial returned error: %v", err)
	}
	defer func(c *websocket.Conn) {
		_ = c.Close()
	}(c)

	// Test sending coordinates
	testCoords := models.ClientCoords{Latitude: 60.1699, Longitude: 24.9384}
	coordsBytes, _ := json.Marshal(testCoords)
	err = c.WriteMessage(websocket.TextMessage, coordsBytes)

	if err != nil {
		t.Fatalf("WriteMessage returned error: %v", err)
	}

	// Read response and validate
	_, message, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage returned error: %v", err)
	}

	var busData models.BusData
	err = json.Unmarshal(message, &busData)
	if err != nil {
		t.Fatalf("Error unmarshalling response: %v", err)
	}

	// Validate response
	if busData.FeedFormat != "gtfsrt" {
		t.Errorf("Expected FeedFormat to be 'gtfsrt', got %s", busData.FeedFormat)
	}

}

func TestHandleGetBusesFromStops(t *testing.T) {
	busChannel := make(chan models.BusData)
	router := mux.NewRouter()
	influxdbClient, err := influxdb.NewBusDataManager()
	if err != nil {
		log.Fatalf("Error connecting to InfluxDB: %v", err)
	}
	defer influxdbClient.GetClient().Close()
	fmt.Println("Connected to InfluxDB")

	busDataSubscriber, _ := mqtt.NewBusDataSubscriber("mqtts://mqtt.digitransit.fi:8883", busChannel)
	busDataService := services.NewBusDataService(influxdbClient, busChannel, busDataSubscriber)
	busHandler := rest.NewBusHandler(busDataService)
	router.HandleFunc("/api/stops/get-busses/", busHandler.HandleGetBusesFromStops).Methods("POST")

	//Mock data
	testData := models.BusData{NextStop: "stop1", VehicleID: "Bus123"}
	busChannel <- testData

	time.Sleep(1 * time.Second)

	server := httptest.NewServer(router)
	defer server.Close()

	stops := []models.BusData{{NextStop: "stop1"}, {NextStop: "stop2"}}
	stopsJSON, _ := json.Marshal(stops)

	resp, err := http.Post(server.URL+"/api/stops/get-busses/", "application/json", bytes.NewBuffer(stopsJSON))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check response content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var busData models.BusData
	if err := json.NewDecoder(resp.Body).Decode(&busData); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	// Check if the busData has valid data, here you might check something more specific
	if busData.VehicleID == "" {
		t.Error("Expected non-empty VehicleID or valid bus data")
	}
}
