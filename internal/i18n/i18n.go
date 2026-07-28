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
		"nav_policies":        "פוליסות",
		"nav_projects":        "פרויקטים",
		"my_projects":         "הפרויקטים שלי",
		"project_name_label":  "שם פרויקט",
		"github_url_label":    "כתובת GitHub",
		"add_project_button":  "הוספת פרויקט",
		"no_projects":         "אין פרויקטים עדיין.",
		"view_project_link":   "צפייה",
		"back_to_projects":    "חזרה לפרויקטים",
		"issues_label":        "Issues",
		"prs_label":           "Pull Requests",
		"commits_label":       "Commits",
		"no_github_url":       "לא הוגדרה כתובת GitHub לפרויקט זה.",
		"github_error_label":  "שגיאה בטעינת נתוני GitHub",
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
		"nav_policies":        "Policies",
		"nav_projects":        "Projects",
		"my_projects":         "My Projects",
		"project_name_label":  "Project name",
		"github_url_label":    "GitHub URL",
		"add_project_button":  "Add project",
		"no_projects":         "No projects yet.",
		"view_project_link":   "View",
		"back_to_projects":    "Back to projects",
		"issues_label":        "Issues",
		"prs_label":           "Pull Requests",
		"commits_label":       "Commits",
		"no_github_url":       "No GitHub URL is set for this project.",
		"github_error_label":  "Error loading GitHub data",
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
