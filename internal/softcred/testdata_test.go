package softcred

import "time"

// Test card identifiers
const (
	cardA = "4567"
	cardB = "8901"
)

// Test merchant names and MCCs
const (
	merchantSoftware    = "ACME SOFTWARE"
	merchantDevTools    = "DEV TOOLS INC"
	merchantCloud       = "WIDGET CLOUD"
	mccSoftware         = "05734"
	mccDevTools         = "07372"
	mccCloudAlternate   = "05045"
	mccPayment          = "00300"
)

// testTx builds a DEBIT transaction with sensible defaults.
// Date defaults to 2026-01-15; override with testTxOn.
func testTx(id, card, merchant, mcc string, amount float64) Transaction {
	return Transaction{
		ID:       id,
		Card:     card,
		Date:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:     "DEBIT",
		Merchant: merchant,
		MCC:      mcc,
		Amount:   amount,
		Memo:     id + "; " + mcc + "; ; ; ;",
	}
}

// testTxOn builds a DEBIT transaction on a specific date.
func testTxOn(id, card, merchant, mcc string, amount float64, date time.Time) Transaction {
	t := testTx(id, card, merchant, mcc, amount)
	t.Date = date
	return t
}

// testCredit builds a CREDIT transaction.
func testCredit(id, card, merchant, mcc string, amount float64, date time.Time) Transaction {
	return Transaction{
		ID:       id,
		Card:     card,
		Date:     date,
		Type:     "CREDIT",
		Merchant: merchant,
		MCC:      mcc,
		Amount:   amount,
		Memo:     id + "; " + mcc + "; ; ; ;",
	}
}

// monthDate returns the 15th of the given year/month.
func monthDate(year int, month time.Month) time.Time {
	return time.Date(year, month, 15, 0, 0, 0, 0, time.UTC)
}
