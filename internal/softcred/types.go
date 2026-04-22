package softcred

import "time"

// Transaction represents a parsed CSV row.
type Transaction struct {
	ID       string
	Card     string
	Date     time.Time
	Type     string // DEBIT or CREDIT
	Merchant string
	MCC      string
	Amount   float64
	Memo     string
}

// FileInfo holds metadata parsed from the CSV filename.
type FileInfo struct {
	Card      string
	StartDate time.Time
	EndDate   time.Time
	Filename  string
}

// ParseResult holds the outcome of parsing a CSV file.
type ParseResult struct {
	Transactions []Transaction
	FileInfo     FileInfo
	Warnings     []string
}

// CardMonth identifies a specific month for a specific card.
type CardMonth struct {
	Card  string
	Month string // "2025-09"
}

// CoverageRecord tracks the date range of an imported file.
type CoverageRecord struct {
	Start    time.Time
	End      time.Time
	Filename string
}

// MCCChange represents a merchant whose MCC has changed.
type MCCChange struct {
	Merchant string
	MCCs     string // comma-separated distinct MCCs
}

// MerchantMonth records a merchant's MCC in a specific month.
type MerchantMonth struct {
	Month string
	MCC   string
}
