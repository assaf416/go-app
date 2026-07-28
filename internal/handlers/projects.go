package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	dbpkg "goapp/internal/db"
	"goapp/internal/githubapi"
	"goapp/internal/i18n"
	"goapp/internal/models"
)

type ProjectsHandler struct {
	DB      *sql.DB
	Fetcher githubapi.Fetcher
}

// NewProjectsHandler builds a ProjectsHandler. Pass a nil fetcher to use
// githubapi.DefaultFetcher (real network calls); tests can pass a stub.
func NewProjectsHandler(conn *sql.DB, fetcher githubapi.Fetcher) *ProjectsHandler {
	if fetcher == nil {
		fetcher = githubapi.DefaultFetcher
	}
	return &ProjectsHandler{DB: conn, Fetcher: fetcher}
}

func (h *ProjectsHandler) currentUser(c echo.Context) (*models.User, error) {
	userID := c.Get("user_id").(int64)
	return dbpkg.FindUserByID(h.DB, userID)
}

func (h *ProjectsHandler) ShowProjects(c echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "user not found")
	}

	projects, err := dbpkg.ListProjects(h.DB)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not load projects")
	}
	lang := i18n.Resolve(c)
	return c.Render(http.StatusOK, "projects.html", map[string]any{
		"User":     user,
		"Projects": projects,
		"Lang":     lang,
		"Dir":      i18n.Dir(lang),
	})
}

func (h *ProjectsHandler) CreateProject(c echo.Context) error {
	name := c.FormValue("project_name")
	githubURL := c.FormValue("github_url")
	if _, err := dbpkg.InsertProject(h.DB, name, githubURL); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not create project")
	}
	return c.Redirect(http.StatusSeeOther, "/projects")
}

func (h *ProjectsHandler) ShowProject(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := dbpkg.FindProjectByID(h.DB, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	user, err := h.currentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "user not found")
	}

	lang := i18n.Resolve(c)
	data := map[string]any{
		"User":    user,
		"Project": project,
		"Lang":    lang,
		"Dir":     i18n.Dir(lang),
	}

	if project.GithubURL != "" {
		activity, fetchErr := h.Fetcher(c.Request().Context(), project.GithubURL)
		if fetchErr != nil {
			data["GithubError"] = fetchErr.Error()
		} else {
			data["Activity"] = activity
		}
	}

	return c.Render(http.StatusOK, "project_detail.html", data)
}
