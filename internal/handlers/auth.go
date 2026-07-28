package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	dbpkg "goapp/internal/db"
	"goapp/internal/i18n"
)

const SessionName = "goapp_session"

type AuthHandler struct {
	DB    *sql.DB
	Store *sessions.CookieStore
}

func NewAuthHandler(conn *sql.DB, store *sessions.CookieStore) *AuthHandler {
	return &AuthHandler{DB: conn, Store: store}
}

func (h *AuthHandler) ShowLogin(c echo.Context) error {
	sess, _ := h.Store.Get(c.Request(), SessionName)
	if _, ok := sess.Values["user_id"]; ok {
		return c.Redirect(http.StatusSeeOther, "/policies")
	}
	lang := i18n.Resolve(c)
	return c.Render(http.StatusOK, "login.html", map[string]any{
		"Error": "",
		"Lang":  lang,
		"Dir":   i18n.Dir(lang),
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	lang := i18n.Resolve(c)

	user, err := dbpkg.FindUserByEmail(h.DB, email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return c.Render(http.StatusUnauthorized, "login_form.html", map[string]any{
			"Error": i18n.T(lang, "invalid_credentials"),
			"Lang":  lang,
			"Dir":   i18n.Dir(lang),
		})
	}

	sess, _ := h.Store.Get(c.Request(), SessionName)
	sess.Values["user_id"] = user.ID
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	c.Response().Header().Set("HX-Redirect", "/policies")
	return c.NoContent(http.StatusOK)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	sess, _ := h.Store.Get(c.Request(), SessionName)
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/login")
}

// RequireAuth is Echo middleware that redirects unauthenticated visitors to /login.
func (h *AuthHandler) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := h.Store.Get(c.Request(), SessionName)
		userID, ok := sess.Values["user_id"]
		if !ok {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		c.Set("user_id", userID)
		return next(c)
	}
}
