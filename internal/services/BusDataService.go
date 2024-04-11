package services

import (
	"finbus/internal/config"
	"finbus/internal/database/influxdb"
	"finbus/internal/models"
	"finbus/internal/transport/mqtt"

	"fmt"
)

type BusDataService interface {
	QueryBusesNear(lat, lon string) ([]models.BusData, error)
	WriteBusData(data models.BusData) error
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
func (s *busDataService) QueryBusesNear(lat, lon string) ([]models.BusData, error) {
	geohash, err := config.ConvertToCustomGeohash(lat, lon)
	if err != nil {
		return nil, err
	}
	return s.influxDBManager.FindBusesNear(geohash)
}

// WriteBusData writes bus telemetry data to the database
func (s *busDataService) WriteBusData(data models.BusData) error {
	return s.influxDBManager.WriteToInfluxDB(data)
}

var _ BusDataService = (*busDataService)(nil)
