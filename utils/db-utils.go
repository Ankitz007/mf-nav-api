package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Ankitz007/mf-nav-api/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func SetupDB() (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUser, DBPassword, DBHost, FundDB)
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %v", err)
	}

	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Minute)

	return db, nil
}

func CheckIfFundExistsInDB(db *sqlx.DB, schemeCode string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
        SELECT 1
        FROM funds
        WHERE scheme_code = ?
    )`

	// Define a context with a timeout for queries
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Execute the query with the context
	err := db.QueryRowContext(ctx, query, schemeCode).Scan(&exists)
	if err != nil {
		exists = false
	}

	return exists, nil
}

func FetchFundDataFromDB(db *sqlx.DB, schemeCode string, startDate, endDate time.Time) (models.JsonResponse, error) {
	var fund models.FundMetadata
	var navRecords []models.NavRecord

	err := db.Get(&fund, "SELECT * FROM funds WHERE scheme_code = ?", schemeCode)
	if err != nil {
		return models.JsonResponse{}, fmt.Errorf("error fetching fund: %v", err)
	}

	// Start with the base query
	query := "SELECT * FROM nav_records WHERE fund_id = ?"
	args := []interface{}{fund.ID}

	// Append date filters to the query if both dates are provided
	if !startDate.IsZero() && !endDate.IsZero() {
		query += " AND date BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	// Add the ordering clause
	query += " ORDER BY date DESC"

	// Fetch the nav records
	err = db.Select(&navRecords, query, args...)
	if err != nil {
		return models.JsonResponse{}, fmt.Errorf("error fetching nav records: %v", err)
	}

	apiResponse := models.JsonResponse{
		Period: fmt.Sprintf("%s to %s", startDate.Format("02-01-2006"), endDate.Format("02-01-2006")),
	}
	apiResponse.Meta.FundHouse = fund.FundHouse
	apiResponse.Meta.SchemeType = fund.SchemeType
	apiResponse.Meta.SchemeCategory = fund.SchemeCategory
	apiResponse.Meta.SchemeCode = int(fund.SchemeCode)
	apiResponse.Meta.SchemeName = fund.SchemeName

	for _, record := range navRecords {
		apiResponse.Data = append(apiResponse.Data, models.NAVData{
			Date: record.Date.Format("02-01-2006"),
			Nav:  fmt.Sprintf("%.4f", record.Nav),
		})
	}

	return apiResponse, nil
}

func WriteDataToDB(wg *sync.WaitGroup, db *sqlx.DB, apiResponse models.JsonResponse, batchSize, concurrencyLimit int) {
	defer wg.Done()

	// Start the transaction
	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("Error starting transaction: %v\n", err)
		return
	}

	// Insert the fund
	fund := models.FundMetadata{
		FundHouse:      apiResponse.Meta.FundHouse,
		SchemeType:     apiResponse.Meta.SchemeType,
		SchemeCategory: apiResponse.Meta.SchemeCategory,
		SchemeCode:     apiResponse.Meta.SchemeCode,
		SchemeName:     apiResponse.Meta.SchemeName,
	}

	fundID, err := insertFundInDBWithTx(tx, fund)
	if err != nil {
		fmt.Printf("Error inserting fund: %v\n", err)
		tx.Rollback() // Rollback the transaction on failure
		return
	}

	var navRecords []models.NavRecord
	for _, record := range apiResponse.Data {
		date, err := time.Parse("02-01-2006", record.Date)
		if err != nil {
			fmt.Printf("Error parsing date %s: %v\n", record.Date, err)
			continue
		}
		nav := models.NavRecord{
			FundID: uint(fundID),
			Date:   date,
			Nav:    AtoF(record.Nav),
		}
		navRecords = append(navRecords, nav)
	}

	// Insert NAV records in batches
	err = insertNavRecordsInDBInBatchesWithTx(tx, navRecords, batchSize, concurrencyLimit)
	if err != nil {
		fmt.Printf("Error inserting NAV records: %v\n", err)
		tx.Rollback() // Rollback if NAV records insertion fails
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		return
	}

	fmt.Printf("Successfully updated data for fund %d in DB\n", apiResponse.Meta.SchemeCode)
}

// insertFundInDBWithTx: Inserts a fund into the DB within a transaction
func insertFundInDBWithTx(tx *sqlx.Tx, fund models.FundMetadata) (int64, error) {
	insertFundQuery := "INSERT INTO funds (fund_house, scheme_type, scheme_category, scheme_code, scheme_name) VALUES (?, ?, ?, ?, ?)"
	result, err := tx.Exec(insertFundQuery, fund.FundHouse, fund.SchemeType, fund.SchemeCategory, fund.SchemeCode, fund.SchemeName)
	if err != nil {
		return 0, fmt.Errorf("error inserting fund: %v", err)
	}

	fundID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error retrieving last insert ID: %v", err)
	}

	return fundID, nil
}

// insertNavRecordsInDBInBatchesWithTx: Inserts NAV records into the DB in batches within a transaction
func insertNavRecordsInDBInBatchesWithTx(tx *sqlx.Tx, navRecords []models.NavRecord, batchSize, concurrencyLimit int) error {
	limitCh := make(chan struct{}, concurrencyLimit)
	var wg_internal sync.WaitGroup

	for i := 0; i < len(navRecords); i += batchSize {
		end := i + batchSize
		if end > len(navRecords) {
			end = len(navRecords)
		}
		batchNavRecords := navRecords[i:end]

		wg_internal.Add(1)
		go func(batch []models.NavRecord) {
			defer wg_internal.Done()
			limitCh <- struct{}{}

			err := insertNavRecordBatchWithTx(tx, batch)
			if err != nil {
				fmt.Printf("Error inserting NAV records batch: %v\n", err)
			}

			<-limitCh
		}(batchNavRecords)
	}

	wg_internal.Wait()

	return nil
}

// insertNavRecordBatchWithTx: Helper function to insert a batch of NAV records within a transaction
func insertNavRecordBatchWithTx(tx *sqlx.Tx, navRecords []models.NavRecord) error {
	insertNavRecordsQuery := "INSERT INTO nav_records (fund_id, date, nav) VALUES "
	navRecordValues := make([]interface{}, 0, len(navRecords)*3)

	for i, record := range navRecords {
		if i > 0 {
			insertNavRecordsQuery += ", "
		}
		insertNavRecordsQuery += "(?, ?, ?)"
		navRecordValues = append(navRecordValues, record.FundID, record.Date, record.Nav)
	}

	_, err := tx.Exec(insertNavRecordsQuery, navRecordValues...)
	if err != nil {
		return fmt.Errorf("error inserting nav_records batch: %v", err)
	}

	return nil
}
