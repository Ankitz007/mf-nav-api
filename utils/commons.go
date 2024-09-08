package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Ankitz007/mf-nav-api/models"
)

// isValidInteger checks if a string can be parsed as an integer.
func isValidInteger(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

func AtoF(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Printf("Error converting string to float: %v\n", err)
		return 0
	}
	return f
}

// validateAndParseDates validates and parses the date strings from the query parameters.
func validateAndParseDates(startDate, endDate string) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if startDate != "" && endDate != "" {
		start, end, err = parseDates(startDate, endDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		if end.After(time.Now()) {
			return time.Time{}, time.Time{}, fmt.Errorf("end date cannot be in the future")
		}

		if start.After(end) {
			return time.Time{}, time.Time{}, fmt.Errorf("start date cannot be after end date")
		}
	} else if startDate == "" && endDate == "" {
		// No dates provided, return all data
		start, end = time.Time{}, time.Time{}
	} else {
		// Only one of the dates provided
		return time.Time{}, time.Time{}, fmt.Errorf("both start and end dates are required in the format dd-mm-yyyy")
	}

	return start, end, nil
}

// parseDates parses the start and end date strings into time.Time objects.
func parseDates(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse("02-01-2006", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format. use dd-mm-yyyy")
	}

	end, err := time.Parse("02-01-2006", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format. use dd-mm-yyyy")
	}

	return start, end, nil
}

func FetchFundData(wg *sync.WaitGroup, mutualFundID string, start, end time.Time) ([]byte, error) {
	var response models.JsonResponse
	var getFromUpstream bool
	var existsInDB bool

	db, err := SetupDB()
	if err != nil {
		// If something is wrong with the DB, fetch data from upstream
		getFromUpstream = true
	}

	if !getFromUpstream {
		existsInDB, err = CheckIfFundExistsInDB(db, mutualFundID)
		if err != nil {
			return nil, fmt.Errorf("error fetching fund from the db")
		}
	}

	if !existsInDB {
		// Fetch fund data from API
		fund, err := fetchFundDataFromUpstream(mutualFundID)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while fetching data from upstream, error %v", err)
		}

		// Check if the meta field is empty, indicating an invalid mutualFundID
		if fund.Meta == (models.JsonResponse{}.Meta) {
			return nil, fmt.Errorf("invalid mutual fund ID")
		}

		// Filter data based on date range
		filteredNAVData := filterNAVDataByDate(fund.Data, start, end)

		response = generateJsonResponseModel(fund.Meta, filteredNAVData, start, end)
		fmt.Printf("Fetched data for fund with ID %s from upstream\n", mutualFundID)

		wg.Add(1)
		go writeDataToDB(db, response, BatchSize, ConcurrencyLimit)
		fmt.Printf("Updated data for fund with id %s in the DB\n", mutualFundID)
	} else {
		// Fetch fund data from the DB
		response, err = FetchFundDataFromDB(db, mutualFundID, start, end)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while fetching data from the db, error %v", err)
		}
		fmt.Printf("Fetched data for fund with ID %s from db\n", mutualFundID)
	}

	// Marshal the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("error creating JSON response")
	}

	return jsonResponse, nil
}
