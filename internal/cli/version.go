package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

// ReadVersion returns the trimmed contents of the version file (version.txt).
func ReadVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// RunVersion prints the application version, styled, to w.
func RunVersion(w io.Writer, versionPath string) error {
	version, err := ReadVersion(versionPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", versionPath, err)
	}
	pterm.DefaultBasicText.WithWriter(w).Printfln(
		"%s %s", pterm.Cyan("goapp"), pterm.Bold.Sprint(version),
	)
	return nil
}
