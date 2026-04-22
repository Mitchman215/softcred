package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"softcred/internal/softcred"
)

func defaultDBPath() string {
	// SOFTCRED_DB env var takes priority
	if p := os.Getenv("SOFTCRED_DB"); p != "" {
		return p
	}
	// XDG_DATA_HOME/tcc/data.db
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "./data.db"
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "softcred", "data.db")
}

func main() {
	dbFlag := flag.String("db", defaultDBPath(), "path to SQLite database")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	// Ensure db directory exists
	if err := os.MkdirAll(filepath.Dir(*dbFlag), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: create db directory: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "import":
		importFlags := flag.NewFlagSet("import", flag.ExitOnError)
		rmFlag := importFlags.Bool("rm", false, "delete CSV files after successful import")
		importFlags.Parse(args[1:])
		importFiles := importFlags.Args()
		if len(importFiles) == 0 {
			fmt.Fprintln(os.Stderr, "error: import requires at least one CSV file")
			fmt.Fprintln(os.Stderr, "usage: softcred import [--rm] <csv-file>...")
			os.Exit(1)
		}
		if err := runImport(*dbFlag, importFiles, *rmFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		db, err := softcred.OpenDB(*dbFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		if err := softcred.RunStatus(db); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "history":
		db, err := softcred.OpenDB(*dbFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		if err := softcred.RunHistory(db); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `softcred — US Bank Triple Cash software credit MCC tracker

Usage:
  softcred [flags] <command> [args...]

Commands:
  import [--rm] <csv-file>...   Import US Bank CSV transaction exports
  status                 Show per-card streak status and warnings
  history                Show per-merchant MCC history over time

Flags:`)
	flag.PrintDefaults()
}

func runImport(dbPath string, files []string, rm bool) error {
	db, err := softcred.OpenDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, path := range files {
		fmt.Printf("Importing %s...\n", path)

		result, err := softcred.ParseCSV(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error: %v (skipping file)\n", err)
			continue
		}

		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  warning: %s\n", w)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		for _, t := range result.Transactions {
			if err := softcred.InsertTransaction(tx, t); err != nil {
				tx.Rollback()
				return fmt.Errorf("insert transaction: %w", err)
			}
		}

		if err := softcred.UpsertCoverage(tx, result.FileInfo.Card, result.FileInfo.Filename, result.FileInfo.StartDate, result.FileInfo.EndDate); err != nil {
			tx.Rollback()
			return fmt.Errorf("record coverage: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %w", err)
		}

		fmt.Printf("  Card %s: %d transactions processed (%s to %s)\n",
			result.FileInfo.Card,
			len(result.Transactions),
			result.FileInfo.StartDate.Format("2006-01-02"),
			result.FileInfo.EndDate.Format("2006-01-02"),
		)

		changes, err := softcred.GetMCCChanges(db, result.FileInfo.Card)
		if err != nil {
			return fmt.Errorf("check MCC changes: %w", err)
		}
		for _, c := range changes {
			fmt.Fprintf(os.Stderr, "  WARNING: %s has multiple MCCs: %s\n", softcred.CleanMerchant(c.Merchant), c.MCCs)
		}

		if rm {
			if err := os.Remove(path); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: failed to delete %s: %v\n", path, err)
			} else {
				fmt.Printf("  Deleted %s\n", path)
			}
		}
	}

	return nil
}
