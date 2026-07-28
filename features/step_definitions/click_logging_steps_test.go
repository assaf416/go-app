package steps

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func (ts *testState) setAppEnv(env string) error {
	ts.appEnvWasSet = true
	return os.Setenv("APP_ENV", env)
}

func (ts *testState) userClicksLoginButton() error {
	body := []byte(`{"path":"/login","tag":"BUTTON","text":"התחברות"}`)
	resp, err := ts.client.Post(ts.server.URL+"/api/log-click", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("expected the click-log request to succeed, got status %d", resp.StatusCode)
	}
	return nil
}

func (ts *testState) clickLoggedInFile(logFile string) error {
	data, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("expected %s to exist: %w", logFile, err)
	}
	if !strings.Contains(string(data), "path=/login") {
		return fmt.Errorf("expected %s to contain the logged click, got:\n%s", logFile, data)
	}
	return nil
}
