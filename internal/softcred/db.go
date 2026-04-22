package softcred

import (
	"database/sql"
	"fmt"
	"time"
)

const schema = `
CREATE TABLE IF NOT EXISTS transactions (
	id TEXT PRIMARY KEY,
	card TEXT NOT NULL,
	date TEXT NOT NULL,
	type TEXT NOT NULL,
	merchant TEXT NOT NULL,
	mcc TEXT,
	amount REAL NOT NULL,
	memo TEXT NOT NULL,
	imported_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%S', 'now'))
);

CREATE TABLE IF NOT EXISTS coverage (
	card TEXT NOT NULL,
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	filename TEXT NOT NULL,
	imported_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%S', 'now')),
	PRIMARY KEY (card, filename)
);

CREATE INDEX IF NOT EXISTS idx_card_date ON transactions(card, date);
CREATE INDEX IF NOT EXISTS idx_merchant_mcc ON transactions(merchant, mcc);
`

// OpenDB opens (or creates) the SQLite database and initializes the schema.
func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}
	return db, nil
}

// Executor is satisfied by both *sql.DB and *sql.Tx.
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// InsertTransaction inserts a transaction, ignoring duplicates by ID.
func InsertTransaction(ex Executor, t Transaction) error {
	_, err := ex.Exec(
		`INSERT OR IGNORE INTO transactions (id, card, date, type, merchant, mcc, amount, memo) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Card, t.Date.Format("2006-01-02"), t.Type, t.Merchant, t.MCC, t.Amount, t.Memo,
	)
	return err
}

// UpsertCoverage records the date range of an imported CSV file.
func UpsertCoverage(ex Executor, card, filename string, start, end time.Time) error {
	_, err := ex.Exec(
		`INSERT INTO coverage (card, start_date, end_date, filename) VALUES (?, ?, ?, ?)
		 ON CONFLICT(card, filename) DO UPDATE SET start_date=excluded.start_date, end_date=excluded.end_date`,
		card, start.Format("2006-01-02"), end.Format("2006-01-02"), filename,
	)
	return err
}

// GetQualifyingPurchases returns all MCC 5734 debit transactions grouped by card+month.
func GetQualifyingPurchases(db *sql.DB) (map[CardMonth][]Transaction, error) {
	rows, err := db.Query(
		`SELECT id, card, date, type, merchant, mcc, amount, memo
		 FROM transactions
		 WHERE type = 'DEBIT' AND mcc = '05734'
		 ORDER BY card, date`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[CardMonth][]Transaction)
	for rows.Next() {
		var t Transaction
		var dateStr string
		if err := rows.Scan(&t.ID, &t.Card, &dateStr, &t.Type, &t.Merchant, &t.MCC, &t.Amount, &t.Memo); err != nil {
			return nil, err
		}
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date %q from db: %w", dateStr, err)
		}
		t.Date = parsed
		key := CardMonth{Card: t.Card, Month: t.Date.Format("2006-01")}
		result[key] = append(result[key], t)
	}
	return result, rows.Err()
}

// GetCards returns all distinct card IDs.
func GetCards(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT card FROM transactions ORDER BY card`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []string
	for rows.Next() {
		var card string
		if err := rows.Scan(&card); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}

// GetCoverage returns all coverage records for a card, ordered by start_date.
func GetCoverage(db *sql.DB, card string) ([]CoverageRecord, error) {
	rows, err := db.Query(
		`SELECT start_date, end_date, filename FROM coverage WHERE card = ? ORDER BY start_date`,
		card,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []CoverageRecord
	for rows.Next() {
		var r CoverageRecord
		var startStr, endStr string
		if err := rows.Scan(&startStr, &endStr, &r.Filename); err != nil {
			return nil, err
		}
		var parseErr error
		r.Start, parseErr = time.Parse("2006-01-02", startStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse coverage start_date %q: %w", startStr, parseErr)
		}
		r.End, parseErr = time.Parse("2006-01-02", endStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse coverage end_date %q: %w", endStr, parseErr)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// GetMerchantMCCHistory returns per-merchant MCC by month for a given card.
func GetMerchantMCCHistory(db *sql.DB, card string) (map[string][]MerchantMonth, error) {
	rows, err := db.Query(
		`SELECT merchant, substr(date, 1, 7) as month, mcc
		 FROM transactions
		 WHERE card = ? AND type = 'DEBIT' AND mcc != ''
		 GROUP BY merchant, month, mcc
		 ORDER BY merchant, month`,
		card,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]MerchantMonth)
	for rows.Next() {
		var merchant, month, mcc string
		if err := rows.Scan(&merchant, &month, &mcc); err != nil {
			return nil, err
		}
		result[merchant] = append(result[merchant], MerchantMonth{Month: month, MCC: mcc})
	}
	return result, rows.Err()
}

// GetNonPaymentCredits returns CREDIT transactions that are NOT regular payments.
// These are candidates for the $100 software statement credit.
func GetNonPaymentCredits(db *sql.DB, card string) ([]Transaction, error) {
	rows, err := db.Query(
		`SELECT id, card, date, type, merchant, mcc, amount, memo
		 FROM transactions
		 WHERE card = ? AND type = 'CREDIT'
		   AND merchant NOT LIKE '%PAYMENT%THANK YOU%'
		   AND merchant NOT LIKE '%INTERNET PAYMENT%'
		 ORDER BY date`,
		card,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var t Transaction
		var dateStr string
		if err := rows.Scan(&t.ID, &t.Card, &dateStr, &t.Type, &t.Merchant, &t.MCC, &t.Amount, &t.Memo); err != nil {
			return nil, err
		}
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date %q from db: %w", dateStr, err)
		}
		t.Date = parsed
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

// GetMCCChanges returns merchants whose MCC has changed across transactions.
func GetMCCChanges(db *sql.DB, card string) ([]MCCChange, error) {
	rows, err := db.Query(
		`SELECT merchant, GROUP_CONCAT(DISTINCT mcc) as mccs
		 FROM transactions
		 WHERE card = ? AND type = 'DEBIT' AND mcc != ''
		 GROUP BY merchant
		 HAVING COUNT(DISTINCT mcc) > 1`,
		card,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []MCCChange
	for rows.Next() {
		var c MCCChange
		if err := rows.Scan(&c.Merchant, &c.MCCs); err != nil {
			return nil, err
		}
		changes = append(changes, c)
	}
	return changes, rows.Err()
}
