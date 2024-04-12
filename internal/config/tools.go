package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// GetEnv returns the value of an environment variable if it exists, otherwise it returns a fallback value
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Environment variable %s not set. Using fallback value %s\n", key, fallback)
	return fallback
}

// GetGeohash Converts a latitude and longitude to a custom geohash format.
func GetGeohash(lat, lon float64) (string, error) {

	latInt, _ := splitFloat(lat)
	lonInt, _ := splitFloat(lon)

	geohashHead := fmt.Sprintf("%d;%d", latInt, lonInt)

	return fmt.Sprintf("/gtfsrt/vp/+/+/+/+/+/+/+/+/+/+/+/%s/+/+/+/+/#", geohashHead), nil
}

// Splits a float into its integer and fractional parts as strings.
func splitFloat(num float64) (int, string) {
	parts := strings.Split(fmt.Sprintf("%.6f", num), ".") // Ensure 6 decimal places
	intPart, _ := strconv.Atoi(parts[0])
	return intPart, parts[1]
}
