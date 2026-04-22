package softcred

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

// RunStatus prints per-card streak status and warnings.
func RunStatus(db *sql.DB) error {
	cards, err := GetCards(db)
	if err != nil {
		return fmt.Errorf("get cards: %w", err)
	}
	if len(cards) == 0 {
		fmt.Println("No transaction data imported yet. Run: softcred import <csv-file>...")
		return nil
	}

	qualifying, err := GetQualifyingPurchases(db)
	if err != nil {
		return fmt.Errorf("get qualifying purchases: %w", err)
	}

	now := time.Now()
	for i, card := range cards {
		if i > 0 {
			fmt.Println()
		}
		if err := printCardStatus(db, card, qualifying, now); err != nil {
			return err
		}
	}
	return nil
}

func printCardStatus(db *sql.DB, card string, qualifying map[CardMonth][]Transaction, now time.Time) error {
	coverage, err := GetCoverage(db, card)
	if err != nil {
		return fmt.Errorf("get coverage for card %s: %w", card, err)
	}

	if len(coverage) == 0 {
		fmt.Printf("Card %s — No coverage data\n", card)
		return nil
	}

	earliest := coverage[0].Start
	latest := coverage[len(coverage)-1].End

	// Check for coverage gaps
	gaps := FindCoverageGaps(coverage)
	if len(gaps) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: Card %s has coverage gaps:\n", card)
		for _, g := range gaps {
			fmt.Fprintf(os.Stderr, "  No data between %s and %s\n", g[0], g[1])
		}
	}

	// Build month list from earliest to now
	if latest.Before(now) {
		latest = now
	}
	months := MonthRange(earliest, latest)

	streak := CurrentStreak(months, card, qualifying)

	fmt.Printf("Card %s — Current Streak: %d consecutive months (need 11)\n", card, streak)

	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("  %-10s  %-6s  %s\n", "Month", "Status", "Qualifying Purchases")
	fmt.Println(strings.Repeat("─", 70))

	for _, month := range months {
		key := CardMonth{Card: card, Month: month}
		txns := qualifying[key]

		status := "✗"
		var details string

		if len(txns) > 0 {
			status = "✓"
			var parts []string
			for _, t := range txns {
				parts = append(parts, fmt.Sprintf("%s ($%.2f)", CleanMerchant(t.Merchant), math.Abs(t.Amount)))
			}
			details = strings.Join(parts, ", ")
		} else {
			monthTime, _ := time.Parse("2006-01", month)
			nowMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			if monthTime.Equal(nowMonth) || monthTime.After(nowMonth) {
				status = "⚠"
				details = "No qualifying purchase yet"
			} else {
				details = "No MCC 5734 purchase"
			}
		}

		fmt.Printf("  %-10s  %-6s  %s\n", month, status, details)
	}
	fmt.Println(strings.Repeat("─", 70))

	// Credit detection
	cs, err := CheckCredit(db, card, months, qualifying, now)
	if err != nil {
		return fmt.Errorf("check credit for card %s: %w", card, err)
	}
	PrintCreditStatus(cs)

	return nil
}

// CurrentStreak counts consecutive months with qualifying purchases,
// ending at the most recent month with data.
func CurrentStreak(months []string, card string, qualifying map[CardMonth][]Transaction) int {
	lastIdx := -1
	for i := len(months) - 1; i >= 0; i-- {
		key := CardMonth{Card: card, Month: months[i]}
		if len(qualifying[key]) > 0 {
			lastIdx = i
			break
		}
	}
	if lastIdx < 0 {
		return 0
	}

	streak := 0
	for i := lastIdx; i >= 0; i-- {
		key := CardMonth{Card: card, Month: months[i]}
		if len(qualifying[key]) > 0 {
			streak++
		} else {
			break
		}
	}
	return streak
}

// MonthRange returns a slice of "YYYY-MM" strings from start to end (inclusive).
func MonthRange(start, end time.Time) []string {
	var months []string
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !current.After(endMonth) {
		months = append(months, current.Format("2006-01"))
		current = current.AddDate(0, 1, 0)
	}
	return months
}

// FindCoverageGaps finds gaps between coverage records.
func FindCoverageGaps(records []CoverageRecord) [][2]string {
	if len(records) < 2 {
		return nil
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Start.Before(records[j].Start)
	})

	var gaps [][2]string
	for i := 1; i < len(records); i++ {
		prevEnd := records[i-1].End
		currStart := records[i].Start
		if currStart.Sub(prevEnd) > 24*time.Hour {
			gaps = append(gaps, [2]string{
				prevEnd.Format("2006-01-02"),
				currStart.Format("2006-01-02"),
			})
		}
	}
	return gaps
}
