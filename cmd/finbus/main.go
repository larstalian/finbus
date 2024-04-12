package main

import (
	"finbus/internal/config"
	"finbus/internal/database/influxdb"
	"finbus/internal/models"
	"finbus/internal/services"
	"finbus/internal/transport/mqtt"
	"finbus/internal/transport/rest"
	"finbus/internal/transport/ws"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// Creates a channel to receive bus data
	dataChannel := make(chan models.BusData)

	// InfluxDB client setup
	influxdbClient, err := influxdb.NewBusDataManager()
	if err != nil {
		log.Fatalf("Error connecting to InfluxDB: %v", err)
	}
	defer influxdbClient.GetClient().Close()
	fmt.Println("Connected to InfluxDB")

	// Initialize the Bus Data Service
	mqttBroker := config.GetEnv("MQTT_BROKER", "mqtts://mqtt.digitransit.fi:8883")

	// Initialize MQTT client and connect to the broker
	mqttClient, err := mqtt.NewBusDataSubscriber(mqttBroker, dataChannel)
	busDataService := services.NewBusDataService(influxdbClient, dataChannel, mqttClient)
	busHandler := rest.NewBusHandler(busDataService)

	webSocketHandler := ws.NewWebSocketHandler(busDataService)

	if err != nil {
		log.Fatalf("Error creating MQTT client: %v", err)
	}

	// Setup HTTP server and routes, passing the bus data service to the REST handler
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Successfully started finbus service\n")
	})

	router.HandleFunc("/api/get-busses", busHandler.HandleQueryBusesNear).Methods("GET")
	router.HandleFunc("/api/stops/get-busses/", busHandler.HandleGetBusesFromStops).Methods("POST")
	router.HandleFunc("/ws/bus-updates", webSocketHandler.HandleBusUpdatesWS)

	// Start the HTTP server
	httpPort := config.GetEnv("HTTP_PORT", "8080")
	log.Printf("Websocket server listening on port %s", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, router))
}
