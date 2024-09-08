package models

import "time"

type FundMetadata struct {
	ID             uint   `db:"id"`
	FundHouse      string `db:"fund_house"`
	SchemeType     string `db:"scheme_type"`
	SchemeCategory string `db:"scheme_category"`
	SchemeCode     int    `db:"scheme_code"`
	SchemeName     string `db:"scheme_name"`
}

type NavRecord struct {
	ID     uint      `db:"id"`
	FundID uint      `db:"fund_id"`
	Date   time.Time `db:"date"`
	Nav    float64   `db:"nav"`
}
