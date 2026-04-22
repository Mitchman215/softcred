# CLAUDE.md

## What This Is

`softcred` is a CLI tool that tracks MCC codes from US Bank Triple Cash credit card CSV exports to ensure software subscriptions keep coding as MCC 5734 (required for the $100 annual software credit).

## Build & Test

```
# Development (requires nix + direnv)
direnv allow
go build -o softcred ./cmd/softcred
go test ./...

# Nix package
nix build .#softcred
```

## Project Structure

```
cmd/softcred/main.go           # CLI entrypoint, flag parsing, command dispatch
internal/softcred/
  types.go                      # Shared types (Transaction, FileInfo, etc.)
  csv.go                        # CSV parsing, filename parsing, MCC extraction
  db.go                         # SQLite schema, queries, inserts
  status.go                     # Streak calculation, month range, coverage gaps
  history.go                    # Per-merchant MCC history display
  credit.go                     # $100 credit detection (milestone, overdue, candidates)
  *_test.go                     # Tests alongside each file
nix/
  package.nix                   # buildGoModule derivation
  hm-module.nix                 # Home Manager module with configurable dataDir
docs/
  usb-triple-cash-research.md   # Domain research from DoC comments, forums, USB terms
```

## Key Design Decisions

- **SQLite via `modernc.org/sqlite`** — pure Go, no CGO. Avoids C toolchain dependency, simpler nix build.
- **All date columns are TEXT** — `modernc.org/sqlite` coerces DATE columns on read, causing parse failures. Store as `YYYY-MM-DD` strings.
- **Dedup by composite ID** — `card|date|memo_ref|amount` prevents duplicates on re-import. The memo reference alone isn't unique (payment rows reuse generic refs like "WEB AUTOMTC").
- **Streak is rolling** — any 11 consecutive months qualifies. Not fixed to calendar year or card anniversary.
- **Credit detection is heuristic** — we don't know the exact description of the $100 credit transaction yet (no card has hit 11 months). Currently matches any non-payment CREDIT of exactly $100 after milestone.

## CSV Format (US Bank export)

```
"Date","Transaction","Name","Memo","Amount"
"2026-04-14","DEBIT","CLAUDE.AI SUBSCRIPTION ANTHROPIC.COM CA","24011346103100126810835; 05734; ; ; ;","-20.00"
```

- Filename: `Credit Card - XXXX_MM-DD-YYYY_MM-DD-YYYY.csv` (card last-4, date range)
- MCC: 2nd semicolon-delimited field in Memo, zero-padded 5 digits (`05734`)
- Only DEBIT rows have meaningful MCC
- CREDIT rows with "PAYMENT THANK YOU" or "INTERNET PAYMENT" are user payments, not the software credit

## Domain Context

The $100 credit requires 11 consecutive months of MCC 5734 purchases. Missing one month resets the counter. Merchants can change MCC codes without notice (Adobe went from reliable to random in 2025). See `docs/usb-triple-cash-research.md` for full research.

## Future Work

- SimpleFIN Bridge integration — auto-detect when charges hit card, notify to download fresh CSV export
- Hardcode $100 credit description once first credit posts on a monitored card
