package steps

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"goapp/internal/config"
)

// The go-app binary is expensive to build, so it's built once and reused
// across every CLI scenario in this test run.
var (
	buildBinOnce sync.Once
	builtBinPath string
	buildBinErr  error
)

func ensureCLIBinaryBuilt() (string, error) {
	buildBinOnce.Do(func() {
		f, err := os.CreateTemp("", "go-app-cli-*")
		if err != nil {
			buildBinErr = err
			return
		}
		f.Close()
		builtBinPath = f.Name()

		cmd := exec.Command("go", "build", "-o", builtBinPath, "./cmd/server")
		cmd.Dir = "../.." // repo root, relative to features/step_definitions
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildBinErr = fmt.Errorf("go build: %w\n%s", err, out)
		}
	})
	return builtBinPath, buildBinErr
}

func (ts *testState) prepareCLIWorkDirIfNeeded() error {
	if ts.workDir != "" {
		return nil
	}
	dir, err := os.MkdirTemp("", "goapp-cli-*")
	if err != nil {
		return err
	}
	ts.workDir = dir
	return nil
}

func (ts *testState) versionFileContains(content string) error {
	if err := ts.prepareCLIWorkDirIfNeeded(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ts.workDir, "version.txt"), []byte(content), 0644)
}

func (ts *testState) runCLICommand(cmdline string) error {
	if err := ts.prepareCLIWorkDirIfNeeded(); err != nil {
		return err
	}
	bin, err := ensureCLIBinaryBuilt()
	if err != nil {
		return err
	}
	args := strings.Fields(strings.TrimPrefix(cmdline, "go-app "))
	cmd := exec.Command(bin, args...)
	cmd.Dir = ts.workDir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	ts.cliRunErr = cmd.Run()
	ts.cliOut = out.String()
	return nil
}

func (ts *testState) cliOutputContains(text string) error {
	if !strings.Contains(ts.cliOut, text) {
		return fmt.Errorf("expected CLI output to contain %q, got:\n%s", text, ts.cliOut)
	}
	return nil
}

func (ts *testState) runSetupCommand(url, username, password string) error {
	if err := ts.prepareCLIWorkDirIfNeeded(); err != nil {
		return err
	}
	bin, err := ensureCLIBinaryBuilt()
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, "--setup")
	cmd.Dir = ts.workDir
	cmd.Stdin = strings.NewReader(url + "\n" + username + "\n" + password + "\n")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	ts.cliRunErr = cmd.Run()
	ts.cliOut = out.String()
	ts.setupURL = url
	ts.setupUser = username
	return nil
}

func (ts *testState) as400FileCreated() error {
	if ts.cliRunErr != nil {
		return fmt.Errorf("go-app --setup failed: %v\noutput:\n%s", ts.cliRunErr, ts.cliOut)
	}
	data, err := os.ReadFile(filepath.Join(ts.workDir, "as400.json"))
	if err != nil {
		return fmt.Errorf("as400.json was not created: %w", err)
	}
	var cfg config.AS400Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if cfg.URL != ts.setupURL {
		return fmt.Errorf("expected url %q, got %q", ts.setupURL, cfg.URL)
	}
	if cfg.Username != ts.setupUser {
		return fmt.Errorf("expected username %q, got %q", ts.setupUser, cfg.Username)
	}
	return nil
}

// --- a minimal in-process SMTP server, just enough to accept one message ---

type smtpCapture struct {
	mu   sync.Mutex
	data []byte
}

func (c *smtpCapture) set(d []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = d
}

func (c *smtpCapture) get() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}

func startFakeSMTP() (addr string, capture *smtpCapture, stop func(), err error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, nil, err
	}
	capture = &smtpCapture{}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleFakeSMTPConn(conn, capture)
		}
	}()
	return ln.Addr().String(), capture, func() { ln.Close() }, nil
}

func handleFakeSMTPConn(conn net.Conn, capture *smtpCapture) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	fmt.Fprint(conn, "220 fake.smtp.local ESMTP\r\n")

	inData := false
	var dataBuf bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		trimmed := strings.TrimRight(line, "\r\n")

		if inData {
			if trimmed == "." {
				inData = false
				capture.set(dataBuf.Bytes())
				fmt.Fprint(conn, "250 OK: queued\r\n")
				continue
			}
			dataBuf.WriteString(trimmed)
			dataBuf.WriteString("\n")
			continue
		}

		switch upper := strings.ToUpper(trimmed); {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			fmt.Fprint(conn, "250 fake.smtp.local\r\n")
		case strings.HasPrefix(upper, "MAIL FROM"), strings.HasPrefix(upper, "RCPT TO"):
			fmt.Fprint(conn, "250 OK\r\n")
		case upper == "DATA":
			inData = true
			dataBuf.Reset()
			fmt.Fprint(conn, "354 End data with <CR><LF>.<CR><LF>\r\n")
		case upper == "QUIT":
			fmt.Fprint(conn, "221 Bye\r\n")
			return
		default:
			fmt.Fprint(conn, "250 OK\r\n")
		}
	}
}

func (ts *testState) smtpConfigured() error {
	if err := ts.prepareCLIWorkDirIfNeeded(); err != nil {
		return err
	}
	addr, capture, stop, err := startFakeSMTP()
	if err != nil {
		return err
	}
	ts.smtpCapture = capture
	ts.smtpStop = stop

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	cfg := config.SMTPConfig{Host: host, Port: port, From: "goapp@example.com", To: []string{"admin@example.com"}}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(ts.workDir, "settings"), 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ts.workDir, "settings", "smtp.json"), data, 0644)
}

func (ts *testState) logFileWithContent() error {
	if err := ts.prepareCLIWorkDirIfNeeded(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ts.workDir, "app.log"), []byte("2026-07-28 12:00:00 test log line\n"), 0644)
}

func (ts *testState) emailWithLogAttachmentSent() error {
	if ts.cliRunErr != nil {
		return fmt.Errorf("go-app --send-log failed: %v\noutput:\n%s", ts.cliRunErr, ts.cliOut)
	}
	if ts.smtpCapture == nil {
		return fmt.Errorf("no fake smtp server was configured for this scenario")
	}
	data := ts.smtpCapture.get()
	if len(data) == 0 {
		return fmt.Errorf("smtp server did not receive any email")
	}
	if !bytes.Contains(data, []byte("app.log")) {
		return fmt.Errorf("expected email to reference app.log, got:\n%s", data)
	}
	if !bytes.Contains(data, []byte("attachment")) {
		return fmt.Errorf("expected email to include an attachment, got:\n%s", data)
	}
	return nil
}
