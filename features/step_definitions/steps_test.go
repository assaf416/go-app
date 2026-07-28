package steps

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"golang.org/x/crypto/bcrypt"

	"goapp/internal/app"
	dbpkg "goapp/internal/db"
)

// testState holds everything a single scenario needs: an isolated sqlite file,
// a running instance of the real app, and an HTTP client that keeps cookies
// (so the login session persists across requests, just like a browser).
type testState struct {
	dbPath   string
	conn     *sql.DB
	server   *httptest.Server
	client   *http.Client
	lastResp *http.Response
	lastBody string
}

func newTestState() *testState {
	f, err := os.CreateTemp("", "goapp-test-*.db")
	if err != nil {
		panic(err)
	}
	f.Close()

	conn := dbpkg.Open(f.Name())
	e := app.New(conn, "test-secret", "../../web/templates/*.html")
	server := httptest.NewServer(e)

	jar, _ := cookiejar.New(nil)
	return &testState{
		dbPath: f.Name(),
		conn:   conn,
		server: server,
		client: &http.Client{Jar: jar},
	}
}

func (ts *testState) close() {
	ts.server.Close()
	ts.conn.Close()
	os.Remove(ts.dbPath)
}

func (ts *testState) userExists(name, email, password string) error {
	_, err := dbpkg.FindUserByEmail(ts.conn, email)
	if err == nil {
		return nil // seeded already, nothing to do
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = ts.conn.Exec(`INSERT INTO users (name, email, password, phone) VALUES (?, ?, ?, ?)`,
		name, email, string(hash), "050-0000000")
	return err
}

func (ts *testState) userIDByName(name string) (int64, error) {
	var id int64
	err := ts.conn.QueryRow(`SELECT id FROM users WHERE name = ?`, name).Scan(&id)
	return id, err
}

func (ts *testState) userHasPolicy(name, company string) error {
	userID, err := ts.userIDByName(name)
	if err != nil {
		return err
	}
	policies, err := dbpkg.PoliciesForUser(ts.conn, userID)
	if err != nil {
		return err
	}
	for _, p := range policies {
		if p.InsuranceCompany == company {
			return nil // seeded already
		}
	}
	_, err = ts.conn.Exec(
		`INSERT INTO policies (user_id, insurance_company, start_date, end_date, coverage, kp_card) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, company, "2024-01-01", "2025-01-01", "כיסוי כללי", "00000000")
	return err
}

func (ts *testState) login(email, password string) error {
	resp, err := ts.client.PostForm(ts.server.URL+"/login", url.Values{
		"email":    {email},
		"password": {password},
	})
	if err != nil {
		return err
	}
	return ts.captureResponse(resp)
}

func (ts *testState) visitPolicies() error {
	resp, err := ts.client.Get(ts.server.URL + "/policies")
	if err != nil {
		return err
	}
	return ts.captureResponse(resp)
}

func (ts *testState) visitLogin() error {
	resp, err := ts.client.Get(ts.server.URL + "/login")
	if err != nil {
		return err
	}
	return ts.captureResponse(resp)
}

func (ts *testState) visitLoginInLang(lang string) error {
	resp, err := ts.client.Get(ts.server.URL + "/login?lang=" + lang)
	if err != nil {
		return err
	}
	return ts.captureResponse(resp)
}

func (ts *testState) dirIs(dir string) error {
	needle := `dir="` + dir + `"`
	if !strings.Contains(ts.lastBody, needle) {
		return fmt.Errorf("expected page to contain %q, got:\n%s", needle, ts.lastBody)
	}
	return nil
}

func (ts *testState) captureResponse(resp *http.Response) error {
	defer resp.Body.Close()
	buf := make([]byte, 0, 8192)
	tmp := make([]byte, 8192)
	for {
		n, err := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	ts.lastResp = resp
	ts.lastBody = string(buf)
	return nil
}

func (ts *testState) loginSucceeded() error {
	if ts.lastResp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected login to succeed with 200, got %d: %s", ts.lastResp.StatusCode, ts.lastBody)
	}
	if ts.lastResp.Header.Get("HX-Redirect") != "/policies" {
		return fmt.Errorf("expected HX-Redirect to /policies, got %q", ts.lastResp.Header.Get("HX-Redirect"))
	}
	return nil
}

func (ts *testState) loginFailed() error {
	if ts.lastResp.StatusCode == http.StatusOK {
		return fmt.Errorf("expected login to fail, but got 200 OK")
	}
	return nil
}

func (ts *testState) bodyContains(text string) error {
	if !strings.Contains(ts.lastBody, text) {
		return fmt.Errorf("expected response body to contain %q, got:\n%s", text, ts.lastBody)
	}
	return nil
}

func (ts *testState) redirectedTo(path string) error {
	if ts.lastResp.Request == nil || ts.lastResp.Request.URL.Path != path {
		return fmt.Errorf("expected final page to be %q, got %q", path, ts.lastResp.Request.URL.Path)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	var ts *testState

	sc.BeforeScenario(func(*godog.Scenario) {
		ts = newTestState()
	})
	sc.AfterScenario(func(*godog.Scenario, error) {
		ts.close()
	})

	sc.Step(`^שהמערכת מכילה את המשתמש "([^"]+)" עם אימייל "([^"]+)" וסיסמה "([^"]+)"$`,
		func(name, email, password string) error { return ts.userExists(name, email, password) })

	sc.Step(`^למשתמש "([^"]+)" יש פוליסה מחברת "([^"]+)"$`,
		func(name, company string) error { return ts.userHasPolicy(name, company) })

	sc.Step(`^המשתמש מתחבר עם אימייל "([^"]+)" וסיסמה "([^"]+)"$`,
		func(email, password string) error { return ts.login(email, password) })

	sc.Step(`^ההתחברות מצליחה$`, func() error { return ts.loginSucceeded() })
	sc.Step(`^ההתחברות נכשלת$`, func() error { return ts.loginFailed() })

	sc.Step(`^מוצג מסך הפוליסות של "([^"]+)"$`, func(name string) error {
		if err := ts.visitPolicies(); err != nil {
			return err
		}
		return ts.bodyContains(name)
	})

	sc.Step(`^הפוליסה מחברת "([^"]+)" מופיעה ברשימה$`,
		func(company string) error { return ts.bodyContains(company) })

	sc.Step(`^מוצגת הודעת שגיאה "([^"]+)"$`,
		func(msg string) error { return ts.bodyContains(msg) })

	sc.Step(`^משתמש לא מחובר מנסה לגשת למסך הפוליסות$`,
		func() error { return ts.visitPolicies() })

	sc.Step(`^המשתמש פותח את מסך ההתחברות$`, func() error { return ts.visitLogin() })

	sc.Step(`^דף ה-HTML כולל הפניה לספריית Bootstrap$`,
		func() error { return ts.bodyContains("bootstrap") })

	sc.Step(`^המשתמש פותח את מסך ההתחברות בשפה "([^"]+)"$`,
		func(lang string) error { return ts.visitLoginInLang(lang) })

	sc.Step(`^כותרת מסך ההתחברות מוצגת בעברית$`, func() error { return ts.bodyContains("ברוכים הבאים") })
	sc.Step(`^כותרת מסך ההתחברות מוצגת באנגלית$`, func() error { return ts.bodyContains("Welcome") })

	sc.Step(`^כיוון הדף הוא מימין לשמאל$`, func() error { return ts.dirIs("rtl") })
	sc.Step(`^כיוון הדף הוא משמאל לימין$`, func() error { return ts.dirIs("ltr") })

	sc.Step(`^הוא מופנה למסך ההתחברות$`,
		func() error { return ts.redirectedTo("/login") })
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{".."},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned from godog, failed to run feature tests")
	}
}
