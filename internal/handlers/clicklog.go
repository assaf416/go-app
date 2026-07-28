package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
)

type ClickLogHandler struct{}

func NewClickLogHandler() *ClickLogHandler {
	return &ClickLogHandler{}
}

type clickPayload struct {
	Path string `json:"path"`
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

// ClickLogPath returns the log file to append clicks to, based on APP_ENV:
// "production" -> logs/production.log, anything else (including unset) ->
// logs/development.log.
func ClickLogPath() string {
	env := os.Getenv("APP_ENV")
	if env != "production" {
		env = "development"
	}
	return filepath.Join("logs", env+".log")
}

// LogClick appends one line describing a user click to the environment's log
// file. It's intentionally lenient (a logging endpoint shouldn't break the
// page): malformed bodies and write failures return an error status but
// never panic.
func (h *ClickLogHandler) LogClick(c echo.Context) error {
	var payload clickPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid click payload")
	}

	path := ClickLogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not prepare log directory")
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not open click log")
	}
	defer f.Close()

	line := time.Now().UTC().Format(time.RFC3339) +
		" click path=" + payload.Path + " tag=" + payload.Tag + " text=" + payload.Text + "\n"
	if _, err := f.WriteString(line); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not write click log")
	}

	return c.NoContent(http.StatusNoContent)
}
