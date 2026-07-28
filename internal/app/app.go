package app

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"goapp/internal/githubapi"
	"goapp/internal/handlers"
)

// New builds the Echo application wired to conn, ready to serve requests or be
// wrapped by httptest.NewServer. templatesGlob points at the html/template files.
// githubFetcher may be nil to use the real githubapi.DefaultFetcher; tests can
// pass a stub to avoid hitting the network.
func New(conn *sql.DB, sessionKey string, templatesGlob string, githubFetcher githubapi.Fetcher) *echo.Echo {
	store := sessions.NewCookieStore([]byte(sessionKey))

	e := echo.New()
	e.Use(middleware.Recover())
	e.Renderer = handlers.NewTemplateRenderer(templatesGlob)

	authHandler := handlers.NewAuthHandler(conn, store)
	policiesHandler := handlers.NewPoliciesHandler(conn)
	projectsHandler := handlers.NewProjectsHandler(conn, githubFetcher)

	e.GET("/", func(c echo.Context) error { return c.Redirect(http.StatusSeeOther, "/login") })
	e.GET("/login", authHandler.ShowLogin)
	e.POST("/login", authHandler.Login)
	e.POST("/logout", authHandler.Logout)

	protected := e.Group("/policies")
	protected.Use(authHandler.RequireAuth)
	protected.GET("", policiesHandler.ShowPolicies)

	projects := e.Group("/projects")
	projects.Use(authHandler.RequireAuth)
	projects.GET("", projectsHandler.ShowProjects)
	projects.POST("", projectsHandler.CreateProject)
	projects.GET("/:id", projectsHandler.ShowProject)

	return e
}
