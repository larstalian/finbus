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

// ConvertToCustomGeohash Converts a latitude and longitude to a custom geohash format.
func ConvertToCustomGeohash(latS, lonS string) (string, error) {
	lat, err := strconv.ParseFloat(latS, 64)
	if err != nil {
		return "", fmt.Errorf("invalid latitude value: %v", err)
	}

	lon, err := strconv.ParseFloat(lonS, 64)
	if err != nil {
		return "", fmt.Errorf("invalid longitude value: %v", err)
	}

	latInt, latFrac := splitFloat(lat)
	lonInt, lonFrac := splitFloat(lon)

	geohashHead := fmt.Sprintf("%d;%d", latInt, lonInt)

	geohashTail := interleaveAndFormat(latFrac, lonFrac)

	return fmt.Sprintf("%s/%s", geohashHead, geohashTail), nil
}

// Splits a float into its integer and fractional parts as strings.
func splitFloat(num float64) (int, string) {
	parts := strings.Split(fmt.Sprintf("%.6f", num), ".") // Ensure 6 decimal places
	intPart, _ := strconv.Atoi(parts[0])
	return intPart, parts[1]
}

// Interleaves digits of two strings and formats them with slashes.
func interleaveAndFormat(a, b string) string {
	var result []string
	for i := 0; i < len(a) || i < len(b); i++ {
		if i < len(a) {
			result = append(result, string(a[i]))
		}
		if i < len(b) {
			result = append(result, string(b[i]))
		}
	}
	return strings.Join(result, "/")
}
