package utils

import "os"

// Define the base URL as a constant
const BaseURL = "https://api.mfapi.in/mf/"
const BatchSize = 1000
const ConcurrencyLimit = 8

var DBHost = os.Getenv("DBHost")
var DBUser = os.Getenv("DBUser")
var DBPassword = os.Getenv("DBPassword")
var FundDB = os.Getenv("FundDB")
