package models

// UpstreamResponse struct to match the upstream API response structure
type UpstreamResponse struct {
	Meta MetaData  `json:"meta"`
	Data []NAVData `json:"data"`
}

// JsonResponse struct for the API response
type JsonResponse struct {
	Meta   MetaData  `json:"meta"`
	Period string    `json:"period,omitempty"`
	Data   []NAVData `json:"data"`
}

// Define a NAVData struct for individual data points
type NAVData struct {
	Date string `json:"date"`
	Nav  string `json:"nav"`
}

type MetaData struct {
	FundHouse      string `json:"fund_house"`
	SchemeType     string `json:"scheme_type"`
	SchemeCategory string `json:"scheme_category"`
	SchemeCode     int    `json:"scheme_code"`
	SchemeName     string `json:"scheme_name"`
}
