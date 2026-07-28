package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/labstack/echo/v4/middleware"
	"github.com/pterm/pterm"

	"goapp/internal/app"
	"goapp/internal/cli"
	"goapp/internal/db"
	"goapp/internal/monitoring"
)

const (
	defaultVersionPath = "version.txt"
	defaultAS400Path   = "as400.json"
	defaultSMTPPath    = "settings/smtp.json"
	defaultLogPath     = "app.log"
)

func main() {
	versionFlag := flag.Bool("version", false, "print the application version and exit")
	setupFlag := flag.Bool("setup", false, "interactively configure the comtec-as400 connection")
	sendLogFlag := flag.Bool("send-log", false, "email the application log file as an attachment")
	flag.Parse()

	sentryEnabled, flushSentry, err := monitoring.InitSentry()
	if err != nil {
		pterm.Warning.Printfln("Sentry did not initialize: %v", err)
	}
	defer flushSentry()

	switch {
	case *versionFlag:
		if err := cli.RunVersion(os.Stdout, defaultVersionPath); err != nil {
			monitoring.CaptureError(err)
			pterm.Error.Println(err)
			os.Exit(1)
		}
		return
	case *setupFlag:
		if _, err := cli.RunSetup(os.Stdin, os.Stdout, defaultAS400Path); err != nil {
			monitoring.CaptureError(err)
			pterm.Error.Println(err)
			os.Exit(1)
		}
		return
	case *sendLogFlag:
		if err := cli.RunSendLog(os.Stdout, defaultLogPath, defaultSMTPPath); err != nil {
			monitoring.CaptureError(err)
			pterm.Error.Println(err)
			os.Exit(1)
		}
		return
	}

	runServer(sentryEnabled)
}

func runServer(sentryEnabled bool) {
	logFile, err := os.OpenFile(defaultLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "goapp.db"
	}

	conn := db.Open(dbPath)
	defer conn.Close()

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		sessionKey = "dev-secret-change-me-in-production"
	}

	e := app.New(conn, sessionKey, "web/templates/*.html", nil)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: io.MultiWriter(os.Stdout, logFile),
	}))
	if sentryEnabled {
		e.Use(monitoring.EchoMiddleware())
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
