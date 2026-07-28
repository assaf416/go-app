package steps

import (
	"fmt"
	"os"

	"goapp/internal/monitoring"
)

func (ts *testState) sentryDSNUnset() error {
	return os.Unsetenv("SENTRY_DSN")
}

func (ts *testState) sentryDSNSetToValid() error {
	return os.Setenv("SENTRY_DSN", "https://examplePublicKey@o0.ingest.sentry.io/0")
}

func (ts *testState) appStarts() error {
	enabled, flush, err := monitoring.InitSentry()
	ts.sentryEnabled = enabled
	ts.sentryInitErr = err
	if flush != nil {
		flush()
	}
	return nil
}

func (ts *testState) appRunsWithoutErrors() error {
	if ts.sentryInitErr != nil {
		return fmt.Errorf("expected the app to start without errors, got: %v", ts.sentryInitErr)
	}
	return nil
}

func (ts *testState) sentryNotInitialized() error {
	if ts.sentryEnabled {
		return fmt.Errorf("expected Sentry reporting to stay disabled, but it was initialized")
	}
	return nil
}

func (ts *testState) sentryInitialized() error {
	if ts.sentryInitErr != nil {
		return fmt.Errorf("expected Sentry to initialize without error, got: %v", ts.sentryInitErr)
	}
	if !ts.sentryEnabled {
		return fmt.Errorf("expected Sentry reporting to be initialized, but it was not")
	}
	return nil
}
