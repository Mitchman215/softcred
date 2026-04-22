package softcred

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// filenamePattern matches: Credit Card - XXXX_MM-DD-YYYY_MM-DD-YYYY.csv
var filenamePattern = regexp.MustCompile(`^Credit Card - (\d{4})_(\d{2}-\d{2}-\d{4})_(\d{2}-\d{2}-\d{4})\.csv$`)

// expectedHeaders defines the required CSV column order.
var expectedHeaders = []string{"Date", "Transaction", "Name", "Memo", "Amount"}

// ParseFilename extracts card ID and date range from the CSV filename.
func ParseFilename(path string) (FileInfo, error) {
	base := filepath.Base(path)
	matches := filenamePattern.FindStringSubmatch(base)
	if matches == nil {
		return FileInfo{}, fmt.Errorf(
			"filename %q does not match expected pattern: Credit Card - XXXX_MM-DD-YYYY_MM-DD-YYYY.csv",
			base,
		)
	}

	start, err := time.Parse("01-02-2006", matches[2])
	if err != nil {
		return FileInfo{}, fmt.Errorf("parse start date %q in filename: %w", matches[2], err)
	}
	end, err := time.Parse("01-02-2006", matches[3])
	if err != nil {
		return FileInfo{}, fmt.Errorf("parse end date %q in filename: %w", matches[3], err)
	}

	return FileInfo{
		Card:      matches[1],
		StartDate: start,
		EndDate:   end,
		Filename:  base,
	}, nil
}

// ParseMemo extracts the transaction reference ID and MCC from the memo field.
// Memo format: "reference; MCC; ; ; ;"
func ParseMemo(memo string) (id string, mcc string, err error) {
	parts := strings.SplitN(memo, ";", 3)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("memo has fewer than 2 semicolon-delimited fields: %q", memo)
	}
	id = strings.TrimSpace(parts[0])
	mcc = strings.TrimSpace(parts[1])
	return id, mcc, nil
}

// ParseCSV reads a US Bank CSV export and returns parsed transactions.
func ParseCSV(path string) (ParseResult, error) {
	var result ParseResult

	fi, err := ParseFilename(path)
	if err != nil {
		return result, err
	}
	result.FileInfo = fi

	f, err := os.Open(path)
	if err != nil {
		return result, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)

	headers, err := reader.Read()
	if err != nil {
		return result, fmt.Errorf("%s: read headers: %w", fi.Filename, err)
	}
	if err := ValidateHeaders(headers); err != nil {
		return result, fmt.Errorf("%s: %w", fi.Filename, err)
	}

	lineNum := 1
	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, fmt.Errorf("%s:%d: read row: %w", fi.Filename, lineNum, err)
		}

		t, warns, err := ParseRow(record, fi.Card, fi.Filename, lineNum)
		result.Warnings = append(result.Warnings, warns...)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		if t != nil {
			result.Transactions = append(result.Transactions, *t)
		}
	}

	if len(result.Transactions) == 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%s: no transactions found", fi.Filename))
	}

	debitCount := 0
	for _, t := range result.Transactions {
		if t.Type == "DEBIT" {
			debitCount++
		}
	}
	if debitCount == 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%s: no DEBIT transactions found", fi.Filename))
	}

	return result, nil
}

// ValidateHeaders checks that CSV headers match the expected format.
func ValidateHeaders(headers []string) error {
	if len(headers) != len(expectedHeaders) {
		return fmt.Errorf("expected %d columns (%s), got %d", len(expectedHeaders), strings.Join(expectedHeaders, ", "), len(headers))
	}
	var missing []string
	for i, expected := range expectedHeaders {
		if headers[i] != expected {
			missing = append(missing, fmt.Sprintf("column %d: expected %q, got %q", i+1, expected, headers[i]))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("header mismatch: %s", strings.Join(missing, "; "))
	}
	return nil
}

// ParseRow parses a single CSV row.
func ParseRow(record []string, card, filename string, lineNum int) (*Transaction, []string, error) {
	if len(record) != 5 {
		return nil, nil, fmt.Errorf("%s:%d: expected 5 fields, got %d", filename, lineNum, len(record))
	}

	date, err := time.Parse("2006-01-02", record[0])
	if err != nil {
		return nil, nil, fmt.Errorf("%s:%d: invalid date %q: %w", filename, lineNum, record[0], err)
	}

	txType := record[1]
	merchant := record[2]
	memo := record[3]

	amount, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return nil, nil, fmt.Errorf("%s:%d: invalid amount %q: %w", filename, lineNum, record[4], err)
	}

	ref, mcc, err := ParseMemo(memo)
	if err != nil {
		return nil, []string{fmt.Sprintf("%s:%d: %v (skipping row)", filename, lineNum, err)}, nil
	}

	var warns []string
	if txType == "DEBIT" && mcc == "" {
		warns = append(warns, fmt.Sprintf("%s:%d: DEBIT row with no MCC for %q — possible format change", filename, lineNum, merchant))
	}

	// Composite ID: card|date|ref|amount — prevents collisions from non-unique memo references
	// (e.g. payment rows with generic "WEB AUTOMTC" refs)
	compositeID := fmt.Sprintf("%s|%s|%s|%s", card, record[0], ref, record[4])

	return &Transaction{
		ID:       compositeID,
		Card:     card,
		Date:     date,
		Type:     txType,
		Merchant: merchant,
		MCC:      mcc,
		Amount:   amount,
		Memo:     memo,
	}, warns, nil
}

// CleanMerchant trims extra whitespace from merchant names.
func CleanMerchant(name string) string {
	fields := strings.Fields(name)
	return strings.Join(fields, " ")
}
