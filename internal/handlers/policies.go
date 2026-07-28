package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	dbpkg "goapp/internal/db"
	"goapp/internal/i18n"
)

type PoliciesHandler struct {
	DB *sql.DB
}

func NewPoliciesHandler(conn *sql.DB) *PoliciesHandler {
	return &PoliciesHandler{DB: conn}
}

func (h *PoliciesHandler) ShowPolicies(c echo.Context) error {
	userID := c.Get("user_id").(int64)

	user, err := dbpkg.FindUserByID(h.DB, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "user not found")
	}

	policies, err := dbpkg.PoliciesForUser(h.DB, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not load policies")
	}

	lang := i18n.Resolve(c)
	return c.Render(http.StatusOK, "policies.html", map[string]any{
		"User":     user,
		"Policies": policies,
		"Lang":     lang,
		"Dir":      i18n.Dir(lang),
	})
}
