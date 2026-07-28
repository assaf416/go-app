package monitoring

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

// InitSentry initializes the Sentry SDK using the SENTRY_DSN environment
// variable. If it's unset, Sentry stays disabled (enabled=false) and the
// rest of the app runs exactly as if this package didn't exist. The
// returned flush func should be deferred so buffered events are sent
// before the process exits.
func InitSentry() (enabled bool, flush func(), err error) {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return false, func() {}, nil
	}
	if err := sentry.Init(sentry.ClientOptions{Dsn: dsn}); err != nil {
		return false, func() {}, err
	}
	return true, func() { sentry.Flush(2 * time.Second) }, nil
}

// CaptureError reports err to Sentry. Safe to call even when Sentry was
// never initialized (sentry-go no-ops in that case).
func CaptureError(err error) {
	if err == nil {
		return
	}
	sentry.CaptureException(err)
}

// EchoMiddleware reports panics and handler errors to Sentry, with a
// request-scoped hub attached to each request's context. It's safe to
// register even when Sentry was never initialized (it just no-ops).
func EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			hub := sentry.CurrentHub().Clone()
			hub.Scope().SetRequest(c.Request())
			ctx := sentry.SetHubOnContext(c.Request().Context(), hub)
			c.SetRequest(c.Request().WithContext(ctx))

			defer func() {
				if r := recover(); r != nil {
					hub.RecoverWithContext(ctx, r)
					panic(r)
				}
			}()

			err := next(c)
			if err != nil {
				hub.CaptureException(err)
			}
			return err
		}
	}
}
