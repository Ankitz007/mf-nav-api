package handler

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/Ankitz007/mf-nav-api/utils"
)

// HTTP handler function to process the request
func Handler(w http.ResponseWriter, r *http.Request) {
	var getFromUpstream bool
	var getFromDB bool
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

	db, err := utils.SetupDB()
	if err != nil {
		// If something is wrong with the DB, fetch data from upstream
		getFromUpstream = true
	}
	defer db.Close()

	if !getFromUpstream {
		getFromDB, err = utils.CheckIfFundExistsInDB(db, mutualFundID)
		if err != nil {
			utils.CreateErrorResponse(w, http.StatusBadRequest, "error fetching fund from the db")
			return
		}
	}

	// Fetch fund data
	fundData, err := utils.FetchFundData(db, mutualFundID, start, end, getFromDB)
	if err != nil {
		utils.CreateErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Filter data based on date range
	filteredNAVData := utils.FilterNAVDataByDate(fundData.Data, start, end)
	filteredFundData := utils.GenerateJsonResponseModel(fundData.Meta, filteredNAVData, start, end)

	jsonResponse, err := json.Marshal(filteredFundData)
	if err != nil {
		utils.CreateErrorResponse(w, http.StatusBadRequest, "error creating JSON response")
		return
	}

	// Return the JSON response immediately
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

	// Update the DB via a separate goroutine
	if !getFromDB {
		wg.Add(1)
		go utils.WriteDataToDB(&wg, db, fundData, utils.BatchSize, utils.ConcurrencyLimit)
	}
	wg.Wait()
}
