package cli

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"goapp/internal/config"
)

// RunSendLog reads the smtp settings from smtpConfigPath and emails the
// contents of logPath as an attachment.
func RunSendLog(out *os.File, logPath, smtpConfigPath string) error {
	spinner, _ := pterm.DefaultSpinner.WithWriter(out).Start("Sending application log via email...")

	cfg, err := config.LoadSMTP(smtpConfigPath)
	if err != nil {
		spinner.Fail("Could not load SMTP settings from " + smtpConfigPath)
		return fmt.Errorf("load smtp config: %w", err)
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		spinner.Fail("Could not read log file " + logPath)
		return fmt.Errorf("read log file: %w", err)
	}

	if err := sendLogEmail(cfg, filepath.Base(logPath), logData); err != nil {
		spinner.Fail("Failed to send log email")
		return err
	}

	spinner.Success(fmt.Sprintf("Sent %s to %s", logPath, strings.Join(cfg.To, ", ")))
	return nil
}

func sendLogEmail(cfg *config.SMTPConfig, filename string, attachment []byte) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	msg := buildLogEmail(cfg, filename, attachment)
	return smtp.SendMail(addr, auth, cfg.From, cfg.To, msg)
}

func buildLogEmail(cfg *config.SMTPConfig, filename string, attachment []byte) []byte {
	const boundary = "goapp-log-boundary"
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "From: %s\r\n", cfg.From)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(cfg.To, ", "))
	fmt.Fprintf(&buf, "Subject: goapp application log\r\n")
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary)

	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(&buf, "Attached is the goapp application log file.\r\n\r\n")

	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	fmt.Fprintf(&buf, "Content-Type: text/plain; name=%s\r\n", filename)
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n")
	fmt.Fprintf(&buf, "Content-Disposition: attachment; filename=%s\r\n\r\n", filename)
	writeBase64Wrapped(&buf, attachment)
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)

	return buf.Bytes()
}

func writeBase64Wrapped(w io.Writer, data []byte) {
	encoded := base64.StdEncoding.EncodeToString(data)
	const lineLen = 76
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		fmt.Fprintf(w, "%s\r\n", encoded[i:end])
	}
}
