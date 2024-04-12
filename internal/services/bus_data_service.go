package services

import (
	"finbus/internal/config"
	"finbus/internal/database/influxdb"
	"finbus/internal/models"
	"finbus/internal/transport/mqtt"

	"fmt"
)

type BusDataService interface {
	QueryBusesNear(lat, lon float64) ([]models.BusData, error)
	WriteBusData(data models.BusData) error
	SubscribeToBusUpdates(coords models.ClientCoords) (chan models.BusData, error)
	GetBusQueryFromStops(stops []models.BusData) (models.BusData, error)
}

type busDataService struct {
	influxDBManager influxdb.BusDataManager
	mqttBroker      mqtt.BusDataSubscriber
	dataChannel     chan models.BusData
}

// NewBusDataService creates a new BusDataService
func NewBusDataService(dbManager influxdb.BusDataManager, dataChannel chan models.BusData, mqttSub mqtt.BusDataSubscriber) BusDataService {
	service := &busDataService{
		influxDBManager: dbManager,
		dataChannel:     dataChannel,
		mqttBroker:      mqttSub,
	}
	go service.processData()
	return service
}

func (s *busDataService) processData() {
	for busData := range s.dataChannel {
		if err := s.WriteBusData(busData); err != nil {
			fmt.Printf("Error processing data: %v\n", err)
		}
	}
}

// QueryBusesNear queries the buses near the specified coordinates
func (s *busDataService) QueryBusesNear(lat, lon float64) ([]models.BusData, error) {
	geohash, err := config.GetGeohash(lat, lon)
	if err != nil {
		return nil, err
	}
	return s.influxDBManager.FindBusesNear(geohash)
}

// WriteBusData writes bus telemetry data to the database
func (s *busDataService) WriteBusData(data models.BusData) error {
	return s.influxDBManager.WriteToInfluxDB(data)
}

func (s *busDataService) SubscribeToBusUpdates(coords models.ClientCoords) (chan models.BusData, error) {
	geohashPosition, err := config.GetGeohash(coords.Latitude, coords.Longitude)
	if err != nil {
		return nil, err
	}

	err = s.mqttBroker.SubscribeToTopic(geohashPosition)
	if err != nil {
		return nil, err
	}

	return s.dataChannel, nil
}

func (s *busDataService) GetBusQueryFromStops(stops []models.BusData) (models.BusData, error) {
	return s.influxDBManager.FindBusesFromStops(stops)
}

var _ BusDataService = (*busDataService)(nil)
