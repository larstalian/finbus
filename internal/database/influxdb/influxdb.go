package influxdb

import (
	"context"
	"finbus/internal/config"
	"finbus/internal/models"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"log"
	"time"
)

var (
	influxURL    = config.GetEnv("INFLUXDB_URL", "http://influxdb:8086")
	influxToken  = config.GetEnv("INFLUXDB_TOKEN", "3gQf-0IZ9vuoXy3mr_ZyPZUtr3mvSHifynt0cmc3c8KFq2yDwzjPLzo2RzuM8OJOoAFsNk3mA-mPtlk7NFAcjQ==")
	influxOrg    = config.GetEnv("INFLUXDB_ORG", "abax")
	influxBucket = config.GetEnv("INFLUXDB_BUCKET", "finbus")
)

type BusDataManager interface {
	GetClient() influxdb2.Client
	WriteToInfluxDB(data models.BusData) error
	QueryData(vehicleID string) ([]models.BusData, error)
	FindBusesNear(geohash string) ([]models.BusData, error)
	FindBusesFromStops(stops []models.BusData) (models.BusData, error)
}

type busDataManager struct {
	client influxdb2.Client
	org    string
	bucket string
}

// NewBusDataManager creates a new InfluxDBClient and connects to InfluxDB
func NewBusDataManager() (BusDataManager, error) {
	client := influxdb2.NewClientWithOptions(influxURL, influxToken, influxdb2.DefaultOptions().SetLogLevel(3))
	_, err := client.Ready(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error connecting to InfluxDB: %v", err)
	}
	return &busDataManager{
		client: client,
		org:    influxOrg,
		bucket: influxBucket,
	}, nil
}

// GetClient returns the InfluxDB client
func (c *busDataManager) GetClient() influxdb2.Client {
	return c.client
}

// WriteToInfluxDB writes bus telemetry data to InfluxDB
func (c *busDataManager) WriteToInfluxDB(data models.BusData) error {
	fmt.Printf("Writing data to InfluxDB: %v\n", data)
	tags := map[string]string{
		"vehicle_id":     data.VehicleID,
		"mode":           data.Mode,
		"route_id":       data.RouteID,
		"trip_id":        data.TripID,
		"trip_headsign":  data.TripHeadsign,
		"next_stop":      data.NextStop,
		"geoHash_head":   data.GeohashHead,
		"geoHash_first":  data.GeohashFirstDeg,
		"geoHash_second": data.GeohashSecondDeg,
		"geoHash_third":  data.GeohashThirdDeg,
	}
	fields := map[string]interface{}{
		"feed_format":  data.FeedFormat,
		"type":         data.Type,
		"feed_id":      data.FeedID,
		"agency_id":    data.AgencyID,
		"agency_name":  data.AgencyName,
		"direction_id": data.DirectionID,
		"start_time":   data.StartTime,
		"short_name":   data.ShortName,
		"color":        data.Color,
	}

	writeAPI := c.client.WriteAPIBlocking(influxOrg, influxBucket)
	point := influxdb2.NewPoint("busTelemetry",
		tags,
		fields,
		time.Now())

	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
		log.Printf("Error writing to InfluxDB: %v", err)
	} else {
		log.Println("Data successfully written to InfluxDB")
	}
	return nil
}

func (c *busDataManager) FindBusesFromStops(stops []models.BusData) (models.BusData, error) {
	var busData models.BusData
	for _, stop := range stops {
		query := fmt.Sprintf(`from(bucket:"%s")
	|> range(start: -1h)
	|> filter(fn: (r) => r._measurement == "busTelemetry" and r.next_stop == "%s")`, influxBucket, stop.NextStop)

		queryAPI := c.client.QueryAPI(influxOrg)

		result, err := queryAPI.Query(context.Background(), query)
		if err != nil {
			return busData, err
		}

		for result.Next() {
			if result.Record().Measurement() == "busTelemetry" {
				busData = models.BusData{
					VehicleID: result.Record().ValueByKey("vehicle_id").(string),
				}
			}
		}

		if result.Err() != nil {
			return busData, result.Err()
		}
	}
	return busData, nil

}

// QueryData queries bus telemetry data from InfluxDB
func (c *busDataManager) QueryData(vehicleID string) ([]models.BusData, error) {
	query := fmt.Sprintf(`from(bucket:"%s")
    |> range(start: -1h)
    |> filter(fn: (r) => r._measurement == "busTelemetry" and r.next_stop == "%s")`, influxBucket, vehicleID)

	queryAPI := c.client.QueryAPI(influxOrg)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var busDataList []models.BusData
	for result.Next() {
		if result.Record().Measurement() == "busTelemetry" {
			busData := models.BusData{
				VehicleID: result.Record().ValueByKey("vehicle_id").(string),
				RouteID:   result.Record().ValueByKey("route_id").(string),
			}
			busDataList = append(busDataList, busData)
		}
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	return busDataList, nil
}

// FindBusesNear queries buses near a specific location
func (c *busDataManager) FindBusesNear(geohash string) ([]models.BusData, error) {
	fluxQuery := fmt.Sprintf(`from(bucket:"%s")
	|> range(start: -1h)
	|> filter(fn: (r) => r._measurement == "busTelemetry" and r.geoHash_head == "%s")`, influxBucket, geohash)

	queryAPI := c.client.QueryAPI(influxOrg)
	result, err := queryAPI.Query(context.Background(), fluxQuery)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	var buses []models.BusData
	for result.Next() {
		bus := models.BusData{
			VehicleID:        result.Record().ValueByKey("vehicle_id").(string),
			RouteID:          result.Record().ValueByKey("route_id").(string),
			GeohashFirstDeg:  result.Record().ValueByKey("geoHash_first").(string),
			GeohashSecondDeg: result.Record().ValueByKey("geoHash_second").(string),
			GeohashThirdDeg:  result.Record().ValueByKey("geoHash_third").(string),
		}
		buses = append(buses, bus)
	}
	if result.Err() != nil {
		log.Printf("Error processing query results: %v", result.Err())
		return nil, result.Err()
	}

	return buses, nil
}

var _ BusDataManager = (*busDataManager)(nil)
