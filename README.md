# softcred

CLI tool for tracking the [US Bank Triple Cash Rewards](https://www.usbank.com/business-banking/business-credit-cards/business-triple-cash-back-credit-card.html) $100 annual software credit.

## The Problem

The Triple Cash card offers a $100 statement credit if you make at least one MCC 5734 purchase every month for 11 consecutive months. Missing a single month resets the counter. Merchants can silently change their MCC coding at any time, and there's no reliable way to know until you check your transaction data.

`softcred` parses US Bank CSV transaction exports, tracks MCC codes per merchant over time, and alerts you when something changes.

## Features

- **Streak tracking** — shows your current consecutive month count toward the 11-month milestone, per card
- **MCC history** — tracks each merchant's MCC over time so you can spot drift
- **MCC change alerts** — warns on import when a merchant's MCC has changed
- **Credit detection** — monitors for the $100 statement credit after you hit the milestone, warns if overdue
- **Coverage gaps** — flags if your imported data has holes
- **Multi-card** — tracks multiple Triple Cash cards independently

## Installation

### Nix Flake

```
nix run github:Mitchman215/softcred
```

### Home Manager

```nix
{
  inputs.softcred.url = "github:Mitchman215/softcred";

  # In your home-manager config:
  imports = [ softcred.homeManagerModules.default ];
  programs.softcred = {
    enable = true;
    # dataDir = "/custom/path";  # optional, defaults to $XDG_DATA_HOME/softcred
  };
}
```

### From Source

Requires Go 1.21+.

```
go install ./cmd/softcred
```

## Usage

### 1. Export transactions from US Bank

Download your credit card transactions as CSV from US Bank online banking. The filename should match the default format:

```
Credit Card - XXXX_MM-DD-YYYY_MM-DD-YYYY.csv
```

where `XXXX` is the last 4 digits of the card.

### 2. Import

```
softcred import "Credit Card - 0273_07-01-2025_04-25-2026.csv"
```

You can import multiple files at once. Duplicate transactions are automatically ignored.

### 3. Check status

```
$ softcred status

Card 0273 — Current Streak: 9 consecutive months (need 11)
──────────────────────────────────────────────────────────────────────
  Month       Status  Qualifying Purchases
──────────────────────────────────────────────────────────────────────
  2025-08     ✓       SIMPLEFIN BRIDGE BRIDGE.SIMPLE GA ($1.50)
  2025-09     ✓       SIMPLEFIN BRIDGE BRIDGE.SIMPLE GA ($1.50)
  ...
  2026-04     ✓       SIMPLEFIN BRIDGE BRIDGE.SIMPLE GA ($1.50)
──────────────────────────────────────────────────────────────────────

  💰 Credit: 2 more months needed to reach 11-month milestone
```

### 4. Check MCC history

```
$ softcred history

Card 3329 — Merchant MCC History
══════════════════════════════════════════════════

CLAUDE.AI SUBSCRIPTION ANTHROPIC.COM CA
  2026-03 to 2026-04: 05734 (consistent)

GITHUB GITHUB.COM CA
  2026-03: 07372
```

## How MCC Checking Works

US Bank CSV exports include the MCC code in the Memo field:

```
"2026-04-14","DEBIT","CLAUDE.AI SUBSCRIPTION ANTHROPIC.COM CA","24011346103100126810835; 05734; ; ; ;","-20.00"
```

The second semicolon-delimited field (`05734`) is the MCC. Only MCC 5734 ("Computer Software Stores") qualifies for the credit.

## Database

Transaction data is stored in a SQLite database. Default location:

| Method | Path |
|--------|------|
| `--db` flag | Whatever you specify |
| `SOFTCRED_DB` env | Whatever you specify |
| XDG default | `$XDG_DATA_HOME/softcred/data.db` |
| Fallback | `~/.local/share/softcred/data.db` |

## Research

See [`docs/usb-triple-cash-research.md`](docs/usb-triple-cash-research.md) for detailed notes on the credit mechanics, known working/non-working services, MCC drift risks, and community data points.
