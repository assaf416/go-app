# go-app

A simple Go web app for managing insurance policies, built on the **HATE** stack:

- **H**TMX — partial page updates without writing JavaScript (used on the login form)
- **A**lpine.js — small bits of client-side reactivity (the policy filter box)
- **T**ailwind CSS — utility-first styling (loaded via the Play CDN, no build step)
- **E**cho — the Go HTTP framework serving everything

Data is stored in **SQLite** (via the pure-Go `modernc.org/sqlite` driver, so no
CGO/gcc is required to build or run). Acceptance tests are written as **Hebrew
Cucumber/Gherkin** features and run with `godog`.

## Features

- Users: name, email, password (bcrypt-hashed), phone number
- Policies per user: insurance company, start/end date, coverage, קופ״ח card number
- Cookie-based session login; the policies page is only visible after login
  and only shows the logged-in user's own policies
- The database is auto-created and seeded with 3 demo users and their
  policies the first time it runs against an empty file
- Projects page: track a list of projects (name + GitHub URL); when a URL is
  set, the project's page shows its recent issues, pull requests, and commits
  straight from the GitHub API

## Project structure

```
cmd/server/main.go            entrypoint: wires config, DB, and the Echo app, starts listening
internal/app/app.go            builds the Echo instance + routes (shared by main.go and tests)
internal/db/db.go               opens the sqlite DB, creates the schema, seeds demo data
internal/db/repo.go             query helpers (find user, list a user's policies)
internal/models/models.go       User / Policy structs
internal/handlers/auth.go       login/logout handlers + the RequireAuth session middleware
internal/handlers/policies.go   the (protected) policies list handler
internal/handlers/renderer.go   html/template renderer adapter for Echo
web/templates/login.html        login page (+ an htmx-swapped partial for login errors)
web/templates/policies.html     policies page (Alpine-powered company filter)
features/policies.feature       Hebrew Gherkin acceptance scenarios
features/step_definitions/      godog step definitions (Go) driving the real app over HTTP
```

## Libraries used

| Library | Purpose |
|---|---|
| [`github.com/labstack/echo/v4`](https://echo.labstack.com/) | HTTP server/router |
| [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) | pure-Go SQLite driver (no CGO) |
| [`github.com/gorilla/sessions`](https://github.com/gorilla/sessions) | signed session cookies |
| [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | password hashing |
| [`github.com/cucumber/godog`](https://github.com/cucumber/godog) | Cucumber/Gherkin (BDD) test runner for Go |
| [`github.com/pterm/pterm`](https://github.com/pterm/pterm) | styled/colored CLI output (`--version`, `--setup`, `--send-log`) |
| [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term) | masked password entry in `--setup` when run in a real terminal |
| [`github.com/getsentry/sentry-go`](https://github.com/getsentry/sentry-go) | error tracking, gated by `SENTRY_DSN` |
| [`github.com/google/go-github`](https://github.com/google/go-github) | GitHub API client for the Projects page (issues/PRs/commits) |
| [htmx](https://htmx.org/) / [Alpine.js](https://alpinejs.dev/) / [Tailwind CSS](https://tailwindcss.com/) | loaded via CDN in the templates, no frontend build step |

## Installation

Requirements: Go 1.22+ (no gcc/CGO needed).

```bash
git clone <this-repo>
cd go-app
go mod download
```

## Running the app

```bash
go run ./cmd/server
```

The server starts on `http://localhost:8080` and creates `goapp.db` in the
working directory on first run, seeded with demo users:

| Email | Password |
|---|---|
| dana@example.com | password123 |
| yossi@example.com | password123 |
| michal@example.com | password123 |

Optional environment variables:

- `PORT` — HTTP port (default `8080`)
- `DB_PATH` — sqlite file path (default `goapp.db`)
- `SESSION_KEY` — key used to sign session cookies (set a real secret in production)
- `SENTRY_DSN` — if set, panics and handler errors (web server and CLI commands)
  are reported to [Sentry](https://sentry.io). Left unset, Sentry is fully
  disabled and has no effect on the app.

## CLI mode

The same binary also runs one-off admin commands instead of starting the web
server, when passed a flag:

```bash
go run ./cmd/server --version     # prints the version from version.txt
go run ./cmd/server --setup       # interactively saves comtec-as400 connection details to ./as400.json
go run ./cmd/server --send-log    # emails ./app.log as an attachment, using ./settings/smtp.json
```

- `--version` reads `version.txt` (repo root) and prints it.
- `--setup` prompts for the comtec-as400 URL, username, and password, then
  writes them to `./as400.json` (0600 permissions; git-ignored — never commit
  real credentials). Password entry is masked when run in a real terminal.
- `--send-log` reads SMTP settings from `./settings/smtp.json` (also
  git-ignored) and emails the contents of `./app.log` as an attachment:
  ```json
  {
    "host": "smtp.example.com",
    "port": 587,
    "username": "smtp-user",
    "password": "smtp-password",
    "from": "goapp@example.com",
    "to": ["ops@example.com"]
  }
  ```
  Leave `username`/`password` empty to send without SMTP authentication.
- Running the server normally (no flags) also appends request logs to
  `./app.log` (in addition to stdout), so there's always something for
  `--send-log` to send.

## Projects & GitHub integration

The **Projects** page (linked from the top nav once logged in) lets you add a
project (name + GitHub URL) and, for any project with a URL set, shows its
most recent issues, pull requests, and commits — read live from the GitHub
API via [`github.com/google/go-github`](https://github.com/google/go-github).

Reading a public repo works with no setup (subject to GitHub's low,
unauthenticated rate limit — 60 requests/hour per IP). For anything beyond
occasional use, or for private repos, set a token:

1. On GitHub: **Settings → Developer settings → Personal access tokens →
   Fine-grained tokens** → **Generate new token**.
2. Scope it to the repo(s) you want to track, with **read-only** permissions
   for `Contents`, `Issues`, `Pull requests`, and `Metadata` — no write access
   is needed.
3. Export it before starting the app:
   ```bash
   export GITHUB_TOKEN=github_pat_...
   go run ./cmd/server
   ```
   (Classic tokens with `public_repo`/`repo` scope work too.)

If a project's GitHub URL can't be reached (bad URL, rate-limited, private
repo without a token), the project page shows a clear error instead of
failing the whole page.

## Running the tests

Hebrew Cucumber acceptance tests (spin up the real app in-process against a
throwaway sqlite file and drive it over HTTP, so no separate server needs to
be running first):

```bash
go test ./features/... -v
```

Everything (unit + feature tests):

```bash
go test ./...
```

## CI

[.github/workflows/ci.yml](.github/workflows/ci.yml) runs `go build`, `go vet`,
and the full test suite (including the Hebrew Cucumber scenarios) on every push
and pull request to `main`.
