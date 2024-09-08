package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func ValidateRequest(mutualFundID, startDate, endDate string) (time.Time, time.Time, error) {
	// Check if mutualFundID is provided and is a valid integer
	if mutualFundID == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("mutualFundID query parameter is required")
	}
	if !isValidInteger(mutualFundID) {
		return time.Time{}, time.Time{}, fmt.Errorf("mutualFundID must be an integer")
	}

	// Validate and parse dates
	start, end, err := validateAndParseDates(startDate, endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return start, end, err
}

// createErrorResponse creates an error response with the given status code and message.
func CreateErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
