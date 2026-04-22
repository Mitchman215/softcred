package softcred

import (
	"fmt"
	"testing"
	"time"
)

func TestFindMilestones(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{}
	var months []string

	for i := 0; i < 11; i++ {
		m := time.Date(2025, time.Month(3+i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		months = append(months, m)
		qualifying[CardMonth{Card: "X", Month: m}] = []Transaction{{ID: "t"}}
	}

	milestones := findMilestones(months, "X", qualifying)
	if len(milestones) != 1 {
		t.Fatalf("expected 1 milestone, got %d", len(milestones))
	}
	if milestones[0] != "2026-01" {
		t.Errorf("milestone = %q, want 2026-01", milestones[0])
	}
}

func TestFindMilestonesWithGap(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{}
	var months []string

	// 5 months, gap, 11 months
	for i := 0; i < 5; i++ {
		m := time.Date(2025, time.Month(1+i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		months = append(months, m)
		qualifying[CardMonth{Card: "X", Month: m}] = []Transaction{{ID: "t"}}
	}
	months = append(months, "2025-06") // gap
	for i := 0; i < 11; i++ {
		m := time.Date(2025, time.Month(7+i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		months = append(months, m)
		qualifying[CardMonth{Card: "X", Month: m}] = []Transaction{{ID: "t"}}
	}

	milestones := findMilestones(months, "X", qualifying)
	if len(milestones) != 1 {
		t.Fatalf("expected 1 milestone, got %d", len(milestones))
	}
	if milestones[0] != "2026-05" {
		t.Errorf("milestone = %q, want 2026-05", milestones[0])
	}
}

func TestFindMilestonesNone(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{}
	var months []string

	for i := 0; i < 10; i++ {
		m := time.Date(2025, time.Month(1+i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		months = append(months, m)
		qualifying[CardMonth{Card: "X", Month: m}] = []Transaction{{ID: "t"}}
	}

	milestones := findMilestones(months, "X", qualifying)
	if len(milestones) != 0 {
		t.Errorf("expected 0 milestones, got %d", len(milestones))
	}
}

func TestFindMilestonesRepeated(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{}
	var months []string

	// 22 consecutive months → milestones at 11 and 22
	for i := 0; i < 22; i++ {
		m := time.Date(2024, time.Month(1+i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		months = append(months, m)
		qualifying[CardMonth{Card: "X", Month: m}] = []Transaction{{ID: "t"}}
	}

	milestones := findMilestones(months, "X", qualifying)
	if len(milestones) != 2 {
		t.Fatalf("expected 2 milestones, got %d: %v", len(milestones), milestones)
	}
	if milestones[0] != "2024-11" {
		t.Errorf("milestone[0] = %q, want 2024-11", milestones[0])
	}
	if milestones[1] != "2025-10" {
		t.Errorf("milestone[1] = %q, want 2025-10", milestones[1])
	}
}

func TestFindMilestonesGapThenRestart(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{}
	var months []string

	// 11 months → milestone 1
	for i := 0; i < 11; i++ {
		m := time.Date(2024, time.Month(1+i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		months = append(months, m)
		qualifying[CardMonth{Card: "X", Month: m}] = []Transaction{{ID: "t"}}
	}
	// Month 12: gap (miss December 2024)
	months = append(months, "2024-12")
	// 11 fresh months → milestone 2
	for i := 0; i < 11; i++ {
		m := time.Date(2025, time.Month(1+i), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		months = append(months, m)
		qualifying[CardMonth{Card: "X", Month: m}] = []Transaction{{ID: "t"}}
	}

	milestones := findMilestones(months, "X", qualifying)
	if len(milestones) != 2 {
		t.Fatalf("expected 2 milestones, got %d: %v", len(milestones), milestones)
	}
	if milestones[0] != "2024-11" {
		t.Errorf("milestone[0] = %q, want 2024-11", milestones[0])
	}
	if milestones[1] != "2025-11" {
		t.Errorf("milestone[1] = %q, want 2025-11", milestones[1])
	}
}

func TestCheckCreditNotYetReached(t *testing.T) {
	db := testDB(t)

	for i := 0; i < 5; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("t%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2025, time.Month(8+i)),
		))
	}

	qualifying, _ := GetQualifyingPurchases(db)
	months := MonthRange(
		time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	)

	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	cs, err := CheckCredit(db, "1234", months, qualifying, now)
	if err != nil {
		t.Fatalf("CheckCredit: %v", err)
	}
	if len(cs.Cycles) != 0 {
		t.Errorf("expected 0 cycles, got %d", len(cs.Cycles))
	}
	if cs.StreakMonths != 5 {
		t.Errorf("streak = %d, want 5", cs.StreakMonths)
	}
}

func TestCheckCreditMilestoneReached(t *testing.T) {
	db := testDB(t)

	for i := 0; i < 11; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("t%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2025, time.Month(3+i)),
		))
	}

	qualifying, _ := GetQualifyingPurchases(db)
	months := MonthRange(
		time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	)

	now := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	cs, err := CheckCredit(db, "1234", months, qualifying, now)
	if err != nil {
		t.Fatalf("CheckCredit: %v", err)
	}
	if len(cs.Cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(cs.Cycles))
	}
	if cs.Cycles[0].MilestoneMonth != "2026-01" {
		t.Errorf("milestone = %q, want 2026-01", cs.Cycles[0].MilestoneMonth)
	}
	if cs.StreakMonths != 0 {
		t.Errorf("streak toward next = %d, want 0 (exactly 11)", cs.StreakMonths)
	}
}

func TestCheckCreditDetected(t *testing.T) {
	db := testDB(t)

	for i := 0; i < 11; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("t%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2025, time.Month(1+i)),
		))
	}

	InsertTransaction(db, testCredit(
		"credit1", "1234", "SOFTWARE CREDIT", "",
		100.00, time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC),
	))

	qualifying, _ := GetQualifyingPurchases(db)
	months := MonthRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	)

	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	cs, err := CheckCredit(db, "1234", months, qualifying, now)
	if err != nil {
		t.Fatalf("CheckCredit: %v", err)
	}
	if len(cs.Cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(cs.Cycles))
	}
	cy := cs.Cycles[0]
	if cy.CreditTxn == nil {
		t.Fatal("expected CreditTxn to be non-nil")
	}
	if cy.CreditTxn.Amount != 100.00 {
		t.Errorf("credit amount = %.2f, want 100.00", cy.CreditTxn.Amount)
	}
}

func TestCheckCreditPaymentsExcluded(t *testing.T) {
	db := testDB(t)

	InsertTransaction(db, testCredit(
		"pay1", "1234", "PAYMENT   THANK YOU", mccPayment,
		100.00, time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
	))

	credits, err := GetNonPaymentCredits(db, "1234")
	if err != nil {
		t.Fatalf("GetNonPaymentCredits: %v", err)
	}
	if len(credits) != 0 {
		t.Errorf("expected 0 non-payment credits, got %d", len(credits))
	}
}

func TestCheckCreditOverdue(t *testing.T) {
	db := testDB(t)

	for i := 0; i < 11; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("t%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2025, time.Month(1+i)),
		))
	}

	qualifying, _ := GetQualifyingPurchases(db)
	months := MonthRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	)

	now := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	cs, err := CheckCredit(db, "1234", months, qualifying, now)
	if err != nil {
		t.Fatalf("CheckCredit: %v", err)
	}
	if len(cs.Cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(cs.Cycles))
	}
	if !cs.Cycles[0].Overdue {
		t.Error("expected Overdue=true")
	}
	if cs.Cycles[0].CreditTxn != nil {
		t.Error("expected no CreditTxn")
	}
}

func TestCheckCreditMultiCycle(t *testing.T) {
	db := testDB(t)

	// 22 consecutive months → 2 milestones
	for i := 0; i < 22; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("t%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2024, time.Month(1+i)),
		))
	}

	// $100 credit for cycle 1 (milestone 2024-11)
	InsertTransaction(db, testCredit(
		"credit1", "1234", "SOFTWARE CREDIT", "",
		100.00, time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
	))
	// $100 credit for cycle 2 (milestone 2025-10)
	InsertTransaction(db, testCredit(
		"credit2", "1234", "SOFTWARE CREDIT", "",
		100.00, time.Date(2025, 12, 5, 0, 0, 0, 0, time.UTC),
	))

	qualifying, _ := GetQualifyingPurchases(db)
	months := MonthRange(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	)

	now := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	cs, err := CheckCredit(db, "1234", months, qualifying, now)
	if err != nil {
		t.Fatalf("CheckCredit: %v", err)
	}
	if len(cs.Cycles) != 2 {
		t.Fatalf("expected 2 cycles, got %d", len(cs.Cycles))
	}

	// Cycle 1: milestone 2024-11, credit found
	if cs.Cycles[0].MilestoneMonth != "2024-11" {
		t.Errorf("cycle 1 milestone = %q, want 2024-11", cs.Cycles[0].MilestoneMonth)
	}
	if cs.Cycles[0].CreditTxn == nil {
		t.Error("cycle 1: expected credit found")
	}
	if cs.Cycles[0].Number != 1 {
		t.Errorf("cycle 1 number = %d, want 1", cs.Cycles[0].Number)
	}

	// Cycle 2: milestone 2025-10, credit found
	if cs.Cycles[1].MilestoneMonth != "2025-10" {
		t.Errorf("cycle 2 milestone = %q, want 2025-10", cs.Cycles[1].MilestoneMonth)
	}
	if cs.Cycles[1].CreditTxn == nil {
		t.Error("cycle 2: expected credit found")
	}
	if cs.Cycles[1].Number != 2 {
		t.Errorf("cycle 2 number = %d, want 2", cs.Cycles[1].Number)
	}

	// No partial streak (22 % 11 = 0)
	if cs.StreakMonths != 0 {
		t.Errorf("streak toward next = %d, want 0", cs.StreakMonths)
	}
}

func TestCheckCreditMultiCyclePartialStreak(t *testing.T) {
	db := testDB(t)

	// 14 consecutive months → 1 milestone + 3 months toward next
	for i := 0; i < 14; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("t%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2025, time.Month(1+i)),
		))
	}

	qualifying, _ := GetQualifyingPurchases(db)
	months := MonthRange(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	)

	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	cs, err := CheckCredit(db, "1234", months, qualifying, now)
	if err != nil {
		t.Fatalf("CheckCredit: %v", err)
	}
	if len(cs.Cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(cs.Cycles))
	}
	if cs.StreakMonths != 3 {
		t.Errorf("streak toward next = %d, want 3", cs.StreakMonths)
	}
}

func TestCheckCreditGapThenRestart(t *testing.T) {
	db := testDB(t)

	// Cycle 1: 11 months (2024-01 to 2024-11)
	for i := 0; i < 11; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("c1_%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2024, time.Month(1+i)),
		))
	}
	// Gap: 2024-12 missing
	// Cycle 2: 11 months (2025-01 to 2025-11)
	for i := 0; i < 11; i++ {
		InsertTransaction(db, testTxOn(
			fmt.Sprintf("c2_%d", i), "1234", merchantSoftware, mccSoftware, -5.00,
			monthDate(2025, time.Month(1+i)),
		))
	}

	// Credit for cycle 1
	InsertTransaction(db, testCredit(
		"credit1", "1234", "SOFTWARE CREDIT", "",
		100.00, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	))

	qualifying, _ := GetQualifyingPurchases(db)
	months := MonthRange(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	)

	// now is past cycle 2's window (2025-11 + 3mo = 2026-02-28)
	now := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	cs, err := CheckCredit(db, "1234", months, qualifying, now)
	if err != nil {
		t.Fatalf("CheckCredit: %v", err)
	}
	if len(cs.Cycles) != 2 {
		t.Fatalf("expected 2 cycles, got %d", len(cs.Cycles))
	}

	// Cycle 1: credited
	if cs.Cycles[0].MilestoneMonth != "2024-11" {
		t.Errorf("cycle 1 milestone = %q, want 2024-11", cs.Cycles[0].MilestoneMonth)
	}
	if cs.Cycles[0].CreditTxn == nil {
		t.Error("cycle 1: expected credit found")
	}

	// Cycle 2: milestone reached after gap restart, overdue (no credit inserted)
	if cs.Cycles[1].MilestoneMonth != "2025-11" {
		t.Errorf("cycle 2 milestone = %q, want 2025-11", cs.Cycles[1].MilestoneMonth)
	}
	if cs.Cycles[1].CreditTxn != nil {
		t.Error("cycle 2: expected no credit yet")
	}
	if !cs.Cycles[1].Overdue {
		t.Error("cycle 2: expected overdue")
	}
}
