package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Ankitz007/mf-nav-api/models"
)

// filterNAVDataByDate filters the NAV data based on the provided date range.
func filterNAVDataByDate(data []models.NAVData, start, end time.Time) []models.NAVData {
	var filteredData []models.NAVData

	for _, item := range data {
		date, err := time.Parse("02-01-2006", item.Date)
		if err != nil {
			continue
		}
		if (start.IsZero() && end.IsZero()) || (date.Equal(start) || date.After(start)) && (date.Equal(end) || date.Before(end)) {
			filteredData = append(filteredData, models.NAVData{Date: item.Date, Nav: item.Nav})
		}
	}

	return filteredData
}

// fetchFundData fetches the fund data from the upstream API using the mutualFundID.
func fetchFundDataFromUpstream(mutualFundID string) (models.UpstreamResponse, error) {
	var fund models.UpstreamResponse

	url := fmt.Sprintf("%s%s", BaseURL, mutualFundID)
	resp, err := http.Get(url)
	if err != nil {
		return fund, fmt.Errorf("error fetching data from upstream API: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&fund); err != nil {
		return fund, fmt.Errorf("error decoding upstream API response: %w", err)
	}

	return fund, nil
}

// generates a successful json response model from upstream response model
func generateJsonResponseModel(meta models.MetaData, data []models.NAVData, start, end time.Time) models.JsonResponse {
	response := models.JsonResponse{
		Meta: meta,
		Data: data,
	}

	if !start.IsZero() && !end.IsZero() {
		response.Period = fmt.Sprintf("%s to %s", start.Format("02-01-2006"), end.Format("02-01-2006"))
	}

	return response
}
