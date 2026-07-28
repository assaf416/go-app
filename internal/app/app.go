package app

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"goapp/internal/handlers"
)

// New builds the Echo application wired to conn, ready to serve requests or be
// wrapped by httptest.NewServer. templatesGlob points at the html/template files.
func New(conn *sql.DB, sessionKey string, templatesGlob string) *echo.Echo {
	store := sessions.NewCookieStore([]byte(sessionKey))

	e := echo.New()
	e.Use(middleware.Recover())
	e.Renderer = handlers.NewTemplateRenderer(templatesGlob)

	authHandler := handlers.NewAuthHandler(conn, store)
	policiesHandler := handlers.NewPoliciesHandler(conn)

	e.GET("/", func(c echo.Context) error { return c.Redirect(http.StatusSeeOther, "/login") })
	e.GET("/login", authHandler.ShowLogin)
	e.POST("/login", authHandler.Login)
	e.POST("/logout", authHandler.Logout)

	protected := e.Group("/policies")
	protected.Use(authHandler.RequireAuth)
	protected.GET("", policiesHandler.ShowPolicies)

	return e
}
