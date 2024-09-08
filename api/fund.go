package handler

import (
	"net/http"
	"sync"

	"github.com/Ankitz007/mf-nav-api/utils"
)

// HTTP handler function to process the request
func Handler(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	// Fetch query parameters
	mutualFundID := r.URL.Query().Get("mutualFundID")
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	// Validate request
	start, end, err := utils.ValidateRequest(mutualFundID, startDate, endDate)
	if err != nil {
		utils.CreateErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Fetch fund data
	jsonResponse, err := utils.FetchFundData(&wg, mutualFundID, start, end)
	if err != nil {
		utils.CreateErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
	wg.Wait()
}
