package softcred

import (
	"testing"
	"time"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		path    string
		want    FileInfo
		wantErr bool
	}{
		{
			path: "data/Credit Card - " + cardA + "_07-01-2025_04-25-2026.csv",
			want: FileInfo{
				Card:      cardA,
				StartDate: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC),
				Filename:  "Credit Card - " + cardA + "_07-01-2025_04-25-2026.csv",
			},
		},
		{
			path: "/some/path/Credit Card - " + cardB + "_08-01-2025_04-25-2026.csv",
			want: FileInfo{
				Card:      cardB,
				StartDate: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC),
				Filename:  "Credit Card - " + cardB + "_08-01-2025_04-25-2026.csv",
			},
		},
		{
			path:    "bad-filename.csv",
			wantErr: true,
		},
		{
			path:    "Credit Card - ABCD_07-01-2025_04-25-2026.csv",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := ParseFilename(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Card != tt.want.Card {
				t.Errorf("Card = %q, want %q", got.Card, tt.want.Card)
			}
			if !got.StartDate.Equal(tt.want.StartDate) {
				t.Errorf("StartDate = %v, want %v", got.StartDate, tt.want.StartDate)
			}
			if !got.EndDate.Equal(tt.want.EndDate) {
				t.Errorf("EndDate = %v, want %v", got.EndDate, tt.want.EndDate)
			}
			if got.Filename != tt.want.Filename {
				t.Errorf("Filename = %q, want %q", got.Filename, tt.want.Filename)
			}
		})
	}
}

func TestParseMemo(t *testing.T) {
	tests := []struct {
		memo    string
		wantID  string
		wantMCC string
		wantErr bool
	}{
		{
			memo:    "10000000000000000000001; 05734; ; ; ;",
			wantID:  "10000000000000000000001",
			wantMCC: "05734",
		},
		{
			memo:    "WEB AUTOMTC; 00300; ; ; ;",
			wantID:  "WEB AUTOMTC",
			wantMCC: "00300",
		},
		{
			memo:    "*W0000; 00300; ; ; ;",
			wantID:  "*W0000",
			wantMCC: "00300",
		},
		{
			memo:    "no-semicolons-here",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.memo, func(t *testing.T) {
			id, mcc, err := ParseMemo(tt.memo)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.wantID {
				t.Errorf("id = %q, want %q", id, tt.wantID)
			}
			if mcc != tt.wantMCC {
				t.Errorf("mcc = %q, want %q", mcc, tt.wantMCC)
			}
		})
	}
}

func TestParseRow(t *testing.T) {
	tests := []struct {
		name    string
		record  []string
		wantID  string
		wantMCC string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "normal debit",
			record:  []string{"2026-04-14", "DEBIT", merchantSoftware, "10000000000000000000001; " + mccSoftware + "; ; ; ;", "-10.00"},
			wantID:  "1234|2026-04-14|10000000000000000000001|-10.00",
			wantMCC: mccSoftware,
		},
		{
			name:    "credit row",
			record:  []string{"2026-04-13", "CREDIT", "PAYMENT THANK YOU", "WEB AUTOMTC; " + mccPayment + "; ; ; ;", "5.00"},
			wantID:  "1234|2026-04-13|WEB AUTOMTC|5.00",
			wantMCC: mccPayment,
		},
		{
			name:    "bad date",
			record:  []string{"not-a-date", "DEBIT", "MERCHANT", "ref; 05734; ; ; ;", "-5.00"},
			wantErr: true,
		},
		{
			name:    "bad amount",
			record:  []string{"2026-01-01", "DEBIT", "MERCHANT", "ref; 05734; ; ; ;", "abc"},
			wantErr: true,
		},
		{
			name:    "wrong field count",
			record:  []string{"2026-01-01", "DEBIT"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, warns, err := ParseRow(tt.record, "1234", "test.csv", 2)
			if tt.wantErr {
				if err == nil && len(warns) == 0 {
					t.Fatalf("expected error or warning, got neither")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil && tx != nil {
				t.Fatalf("expected nil transaction")
			}
			if !tt.wantNil && tx == nil {
				t.Fatalf("expected non-nil transaction")
			}
			if tx != nil && tt.wantID != "" && tx.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", tx.ID, tt.wantID)
			}
			if tx != nil && tx.MCC != tt.wantMCC {
				t.Errorf("MCC = %q, want %q", tx.MCC, tt.wantMCC)
			}
		})
	}
}

func TestValidateHeaders(t *testing.T) {
	if err := ValidateHeaders([]string{"Date", "Transaction", "Name", "Memo", "Amount"}); err != nil {
		t.Errorf("valid headers returned error: %v", err)
	}
	if err := ValidateHeaders([]string{"Date", "Type", "Name", "Memo", "Amount"}); err == nil {
		t.Error("invalid headers returned nil error")
	}
	if err := ValidateHeaders([]string{"Date", "Transaction"}); err == nil {
		t.Error("too few headers returned nil error")
	}
}

func TestCleanMerchant(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"ACME SOFTWARE         ACME.TEST     GA", "ACME SOFTWARE ACME.TEST GA"},
		{"EXAMPLE CO            EXAMPLE.COM   CA", "EXAMPLE CO EXAMPLE.COM CA"},
		{"NORMAL NAME", "NORMAL NAME"},
	}
	for _, tt := range tests {
		got := CleanMerchant(tt.in)
		if got != tt.want {
			t.Errorf("CleanMerchant(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
