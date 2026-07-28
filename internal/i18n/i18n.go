package i18n

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	Hebrew  = "he"
	English = "en"

	CookieName = "lang"
)

var strings = map[string]map[string]string{
	Hebrew: {
		"welcome_title":       "ברוכים הבאים",
		"welcome_subtitle":    "התחברו כדי לצפות בפוליסות הביטוח שלכם",
		"email_label":         "אימייל",
		"password_label":      "סיסמה",
		"login_button":        "התחברות",
		"demo_hint":           "לדוגמה: dana@example.com / password123",
		"invalid_credentials": "אימייל או סיסמה שגויים",
		"logout_button":       "התנתקות",
		"my_policies":         "הפוליסות שלי",
		"filter_placeholder":  "סינון לפי חברת ביטוח...",
		"no_policies":         "אין לך פוליסות פעילות כרגע.",
		"coverage_label":      "כיסוי",
		"kp_card_label":       "מספר כרטיס קופ״ח",
	},
	English: {
		"welcome_title":       "Welcome",
		"welcome_subtitle":    "Log in to view your insurance policies",
		"email_label":         "Email",
		"password_label":      "Password",
		"login_button":        "Log in",
		"demo_hint":           "Demo: dana@example.com / password123",
		"invalid_credentials": "Invalid email or password",
		"logout_button":       "Log out",
		"my_policies":         "My Policies",
		"filter_placeholder":  "Filter by insurance company...",
		"no_policies":         "You have no active policies right now.",
		"coverage_label":      "Coverage",
		"kp_card_label":       "Health-fund card number",
	},
}

// T looks up key in the given language, falling back to Hebrew and then the key itself.
func T(lang, key string) string {
	if m, ok := strings[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := strings[Hebrew][key]; ok {
		return v
	}
	return key
}

// Dir returns the HTML text direction for lang.
func Dir(lang string) string {
	if lang == English {
		return "ltr"
	}
	return "rtl"
}

// Resolve determines the active language for a request: an explicit ?lang=
// query param wins (and is persisted to a cookie for subsequent requests),
// otherwise the lang cookie is used, defaulting to Hebrew.
func Resolve(c echo.Context) string {
	if q := c.QueryParam("lang"); q == Hebrew || q == English {
		c.SetCookie(&http.Cookie{Name: CookieName, Value: q, Path: "/", MaxAge: 60 * 60 * 24 * 365})
		return q
	}
	if ck, err := c.Cookie(CookieName); err == nil && (ck.Value == Hebrew || ck.Value == English) {
		return ck.Value
	}
	return Hebrew
}
