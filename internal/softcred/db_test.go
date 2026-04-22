package softcred

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestInsertAndDedup(t *testing.T) {
	db := testDB(t)
	tx := testTx("ref123", cardA, merchantSoftware, mccSoftware, -5.00)

	if err := InsertTransaction(db, tx); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if err := InsertTransaction(db, tx); err != nil {
		t.Fatalf("duplicate insert: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 row after dedup, got %d", count)
	}
}

func TestGetQualifyingPurchases(t *testing.T) {
	db := testDB(t)

	txns := []Transaction{
		testTxOn("a1", cardA, merchantSoftware, mccSoftware, -5.00, monthDate(2026, 1)),
		testTxOn("a2", cardA, merchantSoftware, mccSoftware, -5.00, monthDate(2026, 2)),
		testTxOn("a3", cardA, merchantDevTools, mccDevTools, -10.00, monthDate(2026, 2)),
		testCredit("a4", cardA, "PAYMENT", mccPayment, 5.00, monthDate(2026, 1)),
	}

	for _, tx := range txns {
		if err := InsertTransaction(db, tx); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	qualifying, err := GetQualifyingPurchases(db)
	if err != nil {
		t.Fatalf("GetQualifyingPurchases: %v", err)
	}

	jan := CardMonth{Card: cardA, Month: "2026-01"}
	if len(qualifying[jan]) != 1 {
		t.Errorf("Jan qualifying = %d, want 1", len(qualifying[jan]))
	}

	feb := CardMonth{Card: cardA, Month: "2026-02"}
	if len(qualifying[feb]) != 1 {
		t.Errorf("Feb qualifying = %d, want 1", len(qualifying[feb]))
	}

	for k, v := range qualifying {
		for _, tx := range v {
			if tx.Type == "CREDIT" {
				t.Errorf("CREDIT transaction in qualifying for %v", k)
			}
		}
	}
}

func TestDateRoundTrip(t *testing.T) {
	db := testDB(t)
	tx := testTxOn("datetest1", "1234", "TEST MERCHANT", mccSoftware, -10.00, monthDate(2026, 3))

	if err := InsertTransaction(db, tx); err != nil {
		t.Fatalf("insert: %v", err)
	}

	qualifying, err := GetQualifyingPurchases(db)
	if err != nil {
		t.Fatalf("GetQualifyingPurchases: %v", err)
	}

	key := CardMonth{Card: "1234", Month: "2026-03"}
	got := qualifying[key]
	if len(got) != 1 {
		t.Fatalf("expected 1 transaction for 2026-03, got %d", len(got))
	}
	if got[0].Date.Year() != 2026 || got[0].Date.Month() != time.March {
		t.Errorf("date roundtrip failed: got %v, want 2026-03-15", got[0].Date)
	}
}

func TestGetCards(t *testing.T) {
	db := testDB(t)

	txns := []Transaction{
		testTxOn("c1", cardB, "M", mccSoftware, -1, monthDate(2026, 1)),
		testTxOn("c2", cardA, "M", mccSoftware, -1, monthDate(2026, 1)),
		testTxOn("c3", cardB, "M", mccSoftware, -1, monthDate(2026, 2)),
	}
	for _, tx := range txns {
		InsertTransaction(db, tx)
	}

	cards, err := GetCards(db)
	if err != nil {
		t.Fatalf("GetCards: %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}
	if cards[0] != cardA || cards[1] != cardB {
		t.Errorf("cards = %v, want [%s %s]", cards, cardA, cardB)
	}
}

func TestMCCChanges(t *testing.T) {
	db := testDB(t)

	txns := []Transaction{
		testTxOn("m1", "1234", merchantCloud, mccSoftware, -5, monthDate(2026, 1)),
		testTxOn("m2", "1234", merchantCloud, mccCloudAlternate, -5, monthDate(2026, 2)),
		testTxOn("m3", "1234", merchantSoftware, mccSoftware, -5, monthDate(2026, 1)),
	}
	for _, tx := range txns {
		InsertTransaction(db, tx)
	}

	changes, err := GetMCCChanges(db, "1234")
	if err != nil {
		t.Fatalf("GetMCCChanges: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 changed merchant, got %d", len(changes))
	}
	if changes[0].Merchant != merchantCloud {
		t.Errorf("changed merchant = %q, want %q", changes[0].Merchant, merchantCloud)
	}
}

func TestCoverage(t *testing.T) {
	db := testDB(t)

	start1 := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	end1 := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	start2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end2 := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)

	UpsertCoverage(db, cardA, "file1.csv", start1, end1)
	UpsertCoverage(db, cardA, "file2.csv", start2, end2)

	records, err := GetCoverage(db, cardA)
	if err != nil {
		t.Fatalf("GetCoverage: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 coverage records, got %d", len(records))
	}
	if !records[0].Start.Equal(start1) {
		t.Errorf("first start = %v, want %v", records[0].Start, start1)
	}
}
