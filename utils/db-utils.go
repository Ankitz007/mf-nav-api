package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/Ankitz007/mutual-funds-nav-backend/models"
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

	err := db.Get(&exists, query, schemeCode)
	if err != nil {
		return false, err
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

	err = db.Select(&navRecords, `
		SELECT * FROM nav_records 
		WHERE fund_id = ? AND date BETWEEN ? AND ? 
		ORDER BY date DESC`, fund.ID, startDate, endDate)
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
			Nav:  fmt.Sprintf("%.2f", record.Nav),
		})
	}

	return apiResponse, nil
}

func writeDataToDB(db *sqlx.DB, apiResponse models.JsonResponse, batchSize, concurrencyLimit int) {
	fund := models.FundMetadata{
		FundHouse:      apiResponse.Meta.FundHouse,
		SchemeType:     apiResponse.Meta.SchemeType,
		SchemeCategory: apiResponse.Meta.SchemeCategory,
		SchemeCode:     apiResponse.Meta.SchemeCode,
		SchemeName:     apiResponse.Meta.SchemeName,
	}

	fundID, err := insertFundInDB(db, fund)
	if err != nil {
		fmt.Printf("Error inserting fund: %v\n", err)
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

	limitCh := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	for i := 0; i < len(navRecords); i += batchSize {
		end := i + batchSize
		if end > len(navRecords) {
			end = len(navRecords)
		}
		batchNavRecords := navRecords[i:end]

		wg.Add(1)
		go insertNavRecordsInDBInBatches(db, batchNavRecords, &wg, limitCh)
	}

	wg.Wait()
}

func insertFundInDB(db *sqlx.DB, fund models.FundMetadata) (int64, error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("error starting transaction for fund: %v", err)
	}

	insertFundQuery := "INSERT INTO funds (fund_house, scheme_type, scheme_category, scheme_code, scheme_name) VALUES (?, ?, ?, ?, ?)"
	result, err := tx.Exec(insertFundQuery, fund.FundHouse, fund.SchemeType, fund.SchemeCategory, fund.SchemeCode, fund.SchemeName)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error inserting fund: %v", err)
	}

	fundID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error retrieving last insert ID: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("error committing transaction for fund: %v", err)
	}

	return fundID, nil
}

func insertNavRecordsInDBInBatches(db *sqlx.DB, navRecords []models.NavRecord, wg *sync.WaitGroup, limitCh chan struct{}) {
	defer wg.Done()

	limitCh <- struct{}{}
	defer func() { <-limitCh }()

	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("Error starting transaction for nav_records batch: %v\n", err)
		return
	}

	insertNavRecordsQuery := "INSERT INTO nav_records (fund_id, date, nav) VALUES "
	navRecordValues := make([]interface{}, 0, len(navRecords)*3)

	for i, record := range navRecords {
		if i > 0 {
			insertNavRecordsQuery += ", "
		}
		insertNavRecordsQuery += "(?, ?, ?)"
		navRecordValues = append(navRecordValues, record.FundID, record.Date, record.Nav)
	}

	stmtNavRecords, err := tx.Preparex(insertNavRecordsQuery)
	if err != nil {
		fmt.Printf("Error preparing statement for nav_records batch: %v\n", err)
		tx.Rollback()
		return
	}
	defer stmtNavRecords.Close()

	_, err = stmtNavRecords.Exec(navRecordValues...)
	if err != nil {
		fmt.Printf("Error inserting nav_records batch: %v\n", err)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("Error committing transaction for nav_records batch: %v\n", err)
		return
	}
}
