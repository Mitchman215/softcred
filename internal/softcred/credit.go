package softcred

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

// CreditCycle represents one 11-month milestone and its corresponding $100 credit.
type CreditCycle struct {
	Number         int          // 1-indexed cycle number
	MilestoneMonth string       // month when 11th consecutive qualifying purchase occurred
	MilestoneDate  time.Time    // last day of the milestone month
	CreditTxn      *Transaction // matched $100 credit (if found)
	Overdue        bool         // true if 2+ billing cycles passed with no credit
	Candidates     []Transaction // non-payment credits in the window that might be the software credit
}

// CreditStatus represents the full credit state for a card across all cycles.
type CreditStatus struct {
	Cycles       []CreditCycle // all completed 11-month milestones
	StreakMonths int           // current partial streak toward next cycle (0-10)
}

// CheckCredit analyzes a card's credit status across all cycles.
func CheckCredit(db *sql.DB, card string, months []string, qualifying map[CardMonth][]Transaction, now time.Time) (CreditStatus, error) {
	var cs CreditStatus

	milestones := findMilestones(months, card, qualifying)
	totalStreak := CurrentStreak(months, card, qualifying)
	cs.StreakMonths = totalStreak % 11

	credits, err := GetNonPaymentCredits(db, card)
	if err != nil {
		return cs, fmt.Errorf("get non-payment credits: %w", err)
	}

	for i, milestone := range milestones {
		cycle := CreditCycle{
			Number:         i + 1,
			MilestoneMonth: milestone,
		}

		milestoneTime, _ := time.Parse("2006-01", milestone)
		cycle.MilestoneDate = time.Date(milestoneTime.Year(), milestoneTime.Month()+1, 0, 0, 0, 0, 0, time.UTC)

		windowStart := cycle.MilestoneDate.AddDate(0, -1, 0)
		windowEnd := cycle.MilestoneDate.AddDate(0, 3, 0)

		for _, c := range credits {
			if c.Date.Before(windowStart) || !c.Date.Before(windowEnd) {
				continue
			}
			cycle.Candidates = append(cycle.Candidates, c)

			if math.Abs(c.Amount-100.0) < 0.01 && cycle.CreditTxn == nil {
				txn := c
				cycle.CreditTxn = &txn
			}
		}

		if cycle.CreditTxn == nil && now.After(windowEnd) {
			cycle.Overdue = true
		}

		cs.Cycles = append(cs.Cycles, cycle)
	}

	return cs, nil
}

// findMilestones returns the months where the 11th consecutive qualifying month was reached.
func findMilestones(months []string, card string, qualifying map[CardMonth][]Transaction) []string {
	var milestones []string
	streak := 0

	for _, month := range months {
		key := CardMonth{Card: card, Month: month}
		if len(qualifying[key]) > 0 {
			streak++
			if streak > 0 && streak%11 == 0 {
				milestones = append(milestones, month)
			}
		} else {
			streak = 0
		}
	}
	return milestones
}

// PrintCreditStatus prints credit detection results for a card.
func PrintCreditStatus(cs CreditStatus) {
	fmt.Println()

	if len(cs.Cycles) == 0 {
		remaining := 11 - cs.StreakMonths
		if remaining > 0 {
			fmt.Printf("  💰 Credit: %d more months needed to reach 11-month milestone\n", remaining)
		}
		return
	}

	fmt.Println("  💰 Credit Cycles:")
	for _, cy := range cs.Cycles {
		prefix := fmt.Sprintf("    Cycle %d: milestone %s —", cy.Number, cy.MilestoneMonth)

		if cy.CreditTxn != nil {
			fmt.Printf("%s ✓ $100 credit on %s (%s)\n",
				prefix,
				cy.CreditTxn.Date.Format("2006-01-02"),
				strings.TrimSpace(CleanMerchant(cy.CreditTxn.Merchant)),
			)
		} else if cy.Overdue {
			fmt.Printf("%s ⚠ OVERDUE — 2+ billing cycles passed, contact US Bank\n", prefix)
		} else {
			fmt.Printf("%s ⏳ expecting by ~%s\n",
				prefix,
				cy.MilestoneDate.AddDate(0, 3, 0).Format("2006-01-02"),
			)
		}

		if len(cy.Candidates) > 0 && cy.CreditTxn == nil {
			for _, c := range cy.Candidates {
				fmt.Printf("      candidate: %s  $%.2f  %s\n",
					c.Date.Format("2006-01-02"),
					c.Amount,
					strings.TrimSpace(CleanMerchant(c.Merchant)),
				)
			}
		}
	}

	if cs.StreakMonths > 0 {
		fmt.Printf("  Current streak: %d months toward next cycle\n", cs.StreakMonths)
	}
}
