package softcred

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// RunHistory prints per-merchant MCC history for all cards.
func RunHistory(db *sql.DB) error {
	cards, err := GetCards(db)
	if err != nil {
		return fmt.Errorf("get cards: %w", err)
	}
	if len(cards) == 0 {
		fmt.Println("No transaction data imported yet. Run: softcred import <csv-file>...")
		return nil
	}

	for i, card := range cards {
		if i > 0 {
			fmt.Println()
		}
		if err := printCardHistory(db, card); err != nil {
			return err
		}
	}
	return nil
}

func printCardHistory(db *sql.DB, card string) error {
	history, err := GetMerchantMCCHistory(db, card)
	if err != nil {
		return fmt.Errorf("get history for card %s: %w", card, err)
	}

	fmt.Printf("Card %s — Merchant MCC History\n", card)
	fmt.Println(strings.Repeat("═", 50))

	merchants := sortedKeys(history)

	for _, merchant := range merchants {
		entries := history[merchant]
		fmt.Printf("\n%s\n", CleanMerchant(merchant))

		allSame := true
		if len(entries) > 1 {
			for _, e := range entries[1:] {
				if e.MCC != entries[0].MCC {
					allSame = false
					break
				}
			}
		}

		if allSame && len(entries) > 1 {
			fmt.Printf("  %s to %s: %s (consistent)\n", entries[0].Month, entries[len(entries)-1].Month, entries[0].MCC)
		} else {
			prevMCC := ""
			for _, e := range entries {
				marker := ""
				if prevMCC != "" && e.MCC != prevMCC {
					marker = " ← CHANGED"
				}
				fmt.Printf("  %s: %s%s\n", e.Month, e.MCC, marker)
				prevMCC = e.MCC
			}
		}
	}

	return nil
}

func sortedKeys(m map[string][]MerchantMonth) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return CleanMerchant(keys[i]) < CleanMerchant(keys[j])
	})
	return keys
}
