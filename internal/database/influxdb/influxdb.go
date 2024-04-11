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
	influxURL    = config.GetEnv("INFLUXDB_URL", "http://localhost:8086")
	influxToken  = config.GetEnv("INFLUXDB_TOKEN", "no-token")
	influxOrg    = config.GetEnv("INFLUXDB_ORG", "abax")
	influxBucket = config.GetEnv("INFLUXDB_BUCKET", "finbus")
)

type BusDataManager interface {
	GetClient() influxdb2.Client
	WriteToInfluxDB(data models.BusData) error
	QueryData(vehicleID string) ([]models.BusData, error)
	FindBusesNear(geohash string) ([]models.BusData, error)
}

type busDataManager struct {
	client influxdb2.Client
	org    string
	bucket string
}

// GetClient returns the InfluxDB client
func (c *busDataManager) GetClient() influxdb2.Client {
	return c.client
}

// NewBusDataManager creates a new InfluxDBClient and connects to InfluxDB
func NewBusDataManager() (BusDataManager, error) {
	client := influxdb2.NewClient(influxURL, influxToken)
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

// WriteToInfluxDB writes bus telemetry data to InfluxDB
func (c *busDataManager) WriteToInfluxDB(data models.BusData) error {
	tags := map[string]string{
		"vehicleID":     data.VehicleID,
		"mode":          data.Mode,
		"routeID":       data.RouteID,
		"geoHashHead":   data.GeohashHead,
		"geoHashFirst":  data.GeohashFirstDeg,
		"geoHashSecond": data.GeohashSecondDeg,
		"geoHashThird":  data.GeohashThirdDeg,
	}
	fields := map[string]interface{}{
		"feed_format":       data.FeedFormat,
		"type":              data.Type,
		"feed_id":           data.FeedID,
		"agency_id":         data.AgencyID,
		"agency_name":       data.AgencyName,
		"direction_id":      data.DirectionID,
		"trip_headsign":     data.TripHeadsign,
		"trip_id":           data.TripID,
		"next_stop":         data.NextStop,
		"start_time":        data.StartTime,
		"geohash_head":      data.GeohashHead,
		"geohash_firstdeg":  data.GeohashFirstDeg,
		"geohash_seconddeg": data.GeohashSecondDeg,
		"geohash_thirddeg":  data.GeohashThirdDeg,
		"short_name":        data.ShortName,
		"color":             data.Color,
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

// QueryData queries bus telemetry data from InfluxDB
func (c *busDataManager) QueryData(vehicleID string) ([]models.BusData, error) {
	query := fmt.Sprintf(`from(bucket:"%s")
    |> range(start: -1h)
    |> filter(fn: (r) => r._measurement == "busTelemetry" and r.vehicleID == "%s")`, influxBucket, vehicleID)

	queryAPI := c.client.QueryAPI(influxOrg)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var busDataList []models.BusData
	for result.Next() {
		if result.Record().Measurement() == "busTelemetry" {
			busData := models.BusData{
				VehicleID: result.Record().ValueByKey("vehicleID").(string),
				Mode:      result.Record().ValueByKey("mode").(string),
				RouteID:   result.Record().ValueByKey("routeID").(string),
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
	|> filter(fn: (r) => r._measurement == "busTelemetry" and r.geoHashHead == "%s")`, influxBucket, geohash)

	queryAPI := c.client.QueryAPI(influxOrg)
	result, err := queryAPI.Query(context.Background(), fluxQuery)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	var buses []models.BusData
	for result.Next() {
		bus := models.BusData{
			VehicleID: result.Record().ValueByKey("vehicleID").(string),
			Mode:      result.Record().ValueByKey("mode").(string),
			RouteID:   result.Record().ValueByKey("routeID").(string),
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
