package lysmail

import (
	"reflect"
	"strings"
	"testing"
)

func validSmtpConfig() SmtpConfig {
	return SmtpConfig{
		Address:        "smtp.example.com:587",
		Host:           "smtp.example.com",
		SenderEmail:    "sender@example.com",
		SenderName:     "Sender",
		SenderPassword: "app-password",
	}
}

func TestBuildEmail_ComposesFields(t *testing.T) {
	cfg := validSmtpConfig()
	cfg.Bccs = []string{"bcc1@example.com"}
	cfg.Ccs = []string{"cfg-cc@example.com"}

	to := []string{"to1@example.com", "to2@example.com"}
	ccs := []string{"arg-cc@example.com"}

	e, err := cfg.buildEmail(to, ccs, "subject", "<b>body</b>")
	if err != nil {
		t.Fatalf("buildEmail returned error: %v", err)
	}

	if e.From != "Sender <sender@example.com>" {
		t.Fatalf("unexpected From: %q", e.From)
	}
	if !reflect.DeepEqual(e.To, to) {
		t.Fatalf("unexpected To: got=%v want=%v", e.To, to)
	}
	if !reflect.DeepEqual(e.Bcc, cfg.Bccs) {
		t.Fatalf("unexpected Bcc: got=%v want=%v", e.Bcc, cfg.Bccs)
	}
	wantCc := []string{"arg-cc@example.com", "cfg-cc@example.com"}
	if !reflect.DeepEqual(e.Cc, wantCc) {
		t.Fatalf("unexpected Cc: got=%v want=%v", e.Cc, wantCc)
	}
	if e.Subject != "subject" {
		t.Fatalf("unexpected Subject: %q", e.Subject)
	}
	if string(e.HTML) != "<b>body</b>" {
		t.Fatalf("unexpected HTML body: %q", string(e.HTML))
	}
}

func TestBuildEmail_RecipientOverride(t *testing.T) {
	cfg := validSmtpConfig()
	cfg.RecipientOverride = "override@example.com"

	e, err := cfg.buildEmail([]string{"to1@example.com"}, nil, "subject", "body")
	if err != nil {
		t.Fatalf("buildEmail returned error: %v", err)
	}

	want := []string{"override@example.com"}
	if !reflect.DeepEqual(e.To, want) {
		t.Fatalf("recipient override not applied: got=%v want=%v", e.To, want)
	}
}

func TestBuildEmail_DefensiveSliceCopies(t *testing.T) {
	cfg := validSmtpConfig()
	cfg.Bccs = []string{"cfg-bcc@example.com"}
	cfg.Ccs = []string{"cfg-cc@example.com"}

	to := make([]string, 1, 4)
	to[0] = "to@example.com"
	ccs := make([]string, 1, 4)
	ccs[0] = "arg-cc@example.com"

	e, err := cfg.buildEmail(to, ccs, "subject", "body")
	if err != nil {
		t.Fatalf("buildEmail returned error: %v", err)
	}

	e.To[0] = "mutated-to@example.com"
	e.Cc[0] = "mutated-cc@example.com"
	e.Bcc[0] = "mutated-bcc@example.com"

	if to[0] != "to@example.com" {
		t.Fatalf("input to slice was mutated: %v", to)
	}
	if ccs[0] != "arg-cc@example.com" {
		t.Fatalf("input ccs slice was mutated: %v", ccs)
	}
	if cfg.Bccs[0] != "cfg-bcc@example.com" {
		t.Fatalf("cfg Bccs slice was mutated: %v", cfg.Bccs)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*SmtpConfig)
		wantErr string
	}{
		{
			name:    "valid config",
			mutate:  func(*SmtpConfig) {},
			wantErr: "",
		},
		{
			name: "missing sender email",
			mutate: func(cfg *SmtpConfig) {
				cfg.SenderEmail = ""
			},
			wantErr: "SenderEmail must be set",
		},
		{
			name: "invalid sender email",
			mutate: func(cfg *SmtpConfig) {
				cfg.SenderEmail = "not-an-email"
			},
			wantErr: "SenderEmail contains invalid email address",
		},
		{
			name: "missing sender password",
			mutate: func(cfg *SmtpConfig) {
				cfg.SenderPassword = ""
			},
			wantErr: "SenderPassword must be set",
		},
		{
			name: "blank bcc",
			mutate: func(cfg *SmtpConfig) {
				cfg.Bccs = []string{""}
			},
			wantErr: "Bccs cannot contain blank email addresses",
		},
		{
			name: "invalid bcc",
			mutate: func(cfg *SmtpConfig) {
				cfg.Bccs = []string{"bad"}
			},
			wantErr: "Bccs contains invalid email address",
		},
		{
			name: "blank cc",
			mutate: func(cfg *SmtpConfig) {
				cfg.Ccs = []string{""}
			},
			wantErr: "Ccs cannot contain blank email addresses",
		},
		{
			name: "invalid cc",
			mutate: func(cfg *SmtpConfig) {
				cfg.Ccs = []string{"bad"}
			},
			wantErr: "Ccs contains invalid email address",
		},
		{
			name: "invalid recipient override",
			mutate: func(cfg *SmtpConfig) {
				cfg.RecipientOverride = "bad"
			},
			wantErr: "RecipientOverride contains invalid email address",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validSmtpConfig()
			tc.mutate(&cfg)

			err := cfg.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q but got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: got=%q want to contain=%q", err.Error(), tc.wantErr)
			}
		})
	}
}
