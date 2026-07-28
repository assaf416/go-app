package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/term"

	"goapp/internal/config"
)

// RunSetup interactively asks for the comtec-as400 connection details and
// saves them to configPath. stdin is an *os.File (rather than a plain
// io.Reader) so that, when it's a real terminal, the password prompt can be
// read without echoing it to the screen; piped input (e.g. in tests) falls
// back to a plain line read.
func RunSetup(stdin *os.File, out *os.File, configPath string) (*config.AS400Config, error) {
	pterm.DefaultHeader.WithFullWidth().WithWriter(out).Println("Comtec AS400 Setup")

	reader := bufio.NewReader(stdin)
	url := promptLine(out, reader, "comtec-as400-url")
	username := promptLine(out, reader, "user")
	password := promptPassword(stdin, out, reader, "password")

	cfg := &config.AS400Config{URL: url, Username: username, Password: password}
	if err := config.SaveAS400(configPath, cfg); err != nil {
		return nil, fmt.Errorf("save %s: %w", configPath, err)
	}

	pterm.Success.WithWriter(out).Printfln("Saved AS400 configuration to %s", configPath)
	return cfg, nil
}

func promptLine(out *os.File, reader *bufio.Reader, label string) string {
	fmt.Fprint(out, pterm.FgCyan.Sprintf("%s: ", label))
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func promptPassword(stdin, out *os.File, reader *bufio.Reader, label string) string {
	fmt.Fprint(out, pterm.FgCyan.Sprintf("%s: ", label))
	if term.IsTerminal(int(stdin.Fd())) {
		b, err := term.ReadPassword(int(stdin.Fd()))
		fmt.Fprintln(out)
		if err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
