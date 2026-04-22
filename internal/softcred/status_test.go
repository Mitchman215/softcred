package softcred

import (
	"testing"
	"time"
)

func TestMonthRange(t *testing.T) {
	start := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)

	months := MonthRange(start, end)
	want := []string{"2025-10", "2025-11", "2025-12", "2026-01", "2026-02"}

	if len(months) != len(want) {
		t.Fatalf("MonthRange length = %d, want %d", len(months), len(want))
	}
	for i, m := range months {
		if m != want[i] {
			t.Errorf("month[%d] = %q, want %q", i, m, want[i])
		}
	}
}

func TestCurrentStreak(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{
		{Card: cardA, Month: "2025-09"}: {{ID: "1"}},
		{Card: cardA, Month: "2025-10"}: {{ID: "2"}},
		{Card: cardA, Month: "2025-11"}: {{ID: "3"}},
		// gap in 2025-12
		{Card: cardA, Month: "2026-01"}: {{ID: "4"}},
		{Card: cardA, Month: "2026-02"}: {{ID: "5"}},
		{Card: cardA, Month: "2026-03"}: {{ID: "6"}},
	}

	months := []string{"2025-09", "2025-10", "2025-11", "2025-12", "2026-01", "2026-02", "2026-03"}

	streak := CurrentStreak(months, cardA, qualifying)
	if streak != 3 {
		t.Errorf("streak = %d, want 3 (Jan-Mar after Dec gap)", streak)
	}
}

func TestCurrentStreakAllQualifying(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{
		{Card: "X", Month: "2025-06"}: {{ID: "1"}},
		{Card: "X", Month: "2025-07"}: {{ID: "2"}},
		{Card: "X", Month: "2025-08"}: {{ID: "3"}},
		{Card: "X", Month: "2025-09"}: {{ID: "4"}},
		{Card: "X", Month: "2025-10"}: {{ID: "5"}},
		{Card: "X", Month: "2025-11"}: {{ID: "6"}},
		{Card: "X", Month: "2025-12"}: {{ID: "7"}},
		{Card: "X", Month: "2026-01"}: {{ID: "8"}},
		{Card: "X", Month: "2026-02"}: {{ID: "9"}},
		{Card: "X", Month: "2026-03"}: {{ID: "10"}},
		{Card: "X", Month: "2026-04"}: {{ID: "11"}},
	}

	months := MonthRange(
		time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	)

	streak := CurrentStreak(months, "X", qualifying)
	if streak != 11 {
		t.Errorf("streak = %d, want 11", streak)
	}
}

func TestCurrentStreakEmpty(t *testing.T) {
	qualifying := map[CardMonth][]Transaction{}
	months := []string{"2026-01", "2026-02"}
	streak := CurrentStreak(months, cardA, qualifying)
	if streak != 0 {
		t.Errorf("streak = %d, want 0", streak)
	}
}

func TestFindCoverageGaps(t *testing.T) {
	records := []CoverageRecord{
		{Start: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)},
		{Start: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
	}

	gaps := FindCoverageGaps(records)
	if len(gaps) != 1 {
		t.Fatalf("expected 1 gap, got %d", len(gaps))
	}
	if gaps[0][0] != "2025-12-31" || gaps[0][1] != "2026-02-01" {
		t.Errorf("gap = %v, want [2025-12-31 2026-02-01]", gaps[0])
	}
}

func TestFindCoverageGapsContiguous(t *testing.T) {
	records := []CoverageRecord{
		{Start: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
	}

	gaps := FindCoverageGaps(records)
	if len(gaps) != 0 {
		t.Errorf("expected no gaps for contiguous records, got %d", len(gaps))
	}
}
