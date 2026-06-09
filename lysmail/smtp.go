package lysmail

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"slices"

	"github.com/jordan-wright/email"
)

type SmtpConfig struct {
	Address           string   // SMTP server address, e.g. "smtp.gmail.com:587"
	Bccs              []string // list of email addresses to be BCC'd on all mails
	Ccs               []string // list of email addresses to be CC'd on all mails
	Host              string   // SMTP server host, e.g. "smtp.gmail.com", used for authentication
	RecipientOverride string   // if non-blank, all emails will be sent to this recipient
	SenderEmail       string   // email of account which sends the mail
	SenderName        string   // name which appears in the recipient's inbox
	SenderPassword    string   // password of account which sends the mail
}

func (cfg *SmtpConfig) buildEmail(to []string, ccs []string, subject string, htmlBody string) (*email.Email, error) {

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", cfg.SenderName, cfg.SenderEmail)

	// use defensive copies of slices to avoid mutating caller's or cfg's slice backing arrays

	finalTo := slices.Clone(to)

	// if RecipientOverride is set, send all emails to that address instead of the supplied 'to' list
	if cfg.RecipientOverride != "" {
		finalTo = []string{cfg.RecipientOverride}
	}
	e.To = finalTo

	e.Bcc = slices.Clone(cfg.Bccs)

	e.Cc = slices.Clone(ccs)
	e.Cc = append(e.Cc, cfg.Ccs...)

	e.Subject = subject
	e.HTML = []byte(htmlBody)

	return e, nil
}

func (cfg *SmtpConfig) Send(to []string, ccs []string, subject string, htmlBody string) (err error) {

	e, err := cfg.buildEmail(to, ccs, subject, htmlBody)
	if err != nil {
		return fmt.Errorf("cfg.buildEmail failed: %w", err)
	}

	err = e.Send(cfg.Address, smtp.PlainAuth("", cfg.SenderEmail, cfg.SenderPassword, cfg.Host))
	if err != nil {
		return fmt.Errorf("e.Send failed: %w", err)
	}

	return nil
}

func (cfg *SmtpConfig) Validate() error {

	// mandatory fields

	// SenderEmail
	if cfg.SenderEmail == "" {
		return fmt.Errorf("SenderEmail must be set")
	}
	_, err := mail.ParseAddress(cfg.SenderEmail)
	if err != nil {
		return fmt.Errorf("SenderEmail contains invalid email address: %s", cfg.SenderEmail)
	}

	// SenderPassword
	if cfg.SenderPassword == "" {
		return fmt.Errorf("SenderPassword must be set")
	}

	// optional fields

	// Bccs
	for _, bcc := range cfg.Bccs {
		if bcc == "" {
			return fmt.Errorf("Bccs cannot contain blank email addresses")
		}
		_, err := mail.ParseAddress(bcc)
		if err != nil {
			return fmt.Errorf("Bccs contains invalid email address: %s", bcc)
		}
	}

	// Ccs
	for _, cc := range cfg.Ccs {
		if cc == "" {
			return fmt.Errorf("Ccs cannot contain blank email addresses")
		}
		_, err := mail.ParseAddress(cc)
		if err != nil {
			return fmt.Errorf("Ccs contains invalid email address: %s", cc)
		}
	}

	// RecipientOverride
	if cfg.RecipientOverride != "" {
		_, err := mail.ParseAddress(cfg.RecipientOverride)
		if err != nil {
			return fmt.Errorf("RecipientOverride contains invalid email address: %s", cfg.RecipientOverride)
		}
	}

	return nil
}
