package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
	"golang.org/x/crypto/bcrypt"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	phone TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS policies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	insurance_company TEXT NOT NULL,
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	coverage TEXT NOT NULL,
	kp_card TEXT NOT NULL
);
`

// Open opens (and if needed creates + seeds) the sqlite database at path.
func Open(path string) *sql.DB {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if _, err := conn.Exec(schema); err != nil {
		log.Fatalf("create schema: %v", err)
	}
	seed(conn)
	return conn
}

// seed inserts demo users + policies the first time the app runs against an empty DB.
func seed(conn *sql.DB) {
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		log.Fatalf("count users: %v", err)
	}
	if count > 0 {
		return
	}

	type seedPolicy struct {
		company, start, end, coverage, kpCard string
	}
	type seedUser struct {
		name, email, password, phone string
		policies                     []seedPolicy
	}

	users := []seedUser{
		{
			name: "דנה כהן", email: "dana@example.com", password: "password123", phone: "050-1234567",
			policies: []seedPolicy{
				{"הראל ביטוח", "2024-01-01", "2025-01-01", "ביטוח בריאות מקיף", "12345678"},
				{"כלל ביטוח", "2023-06-15", "2024-06-15", "ביטוח רכב", "87654321"},
			},
		},
		{
			name: "יוסי לוי", email: "yossi@example.com", password: "password123", phone: "052-7654321",
			policies: []seedPolicy{
				{"מגדל ביטוח", "2024-03-01", "2025-03-01", "ביטוח דירה", "11223344"},
			},
		},
		{
			name: "מיכל אברהם", email: "michal@example.com", password: "password123", phone: "054-9998877",
			policies: []seedPolicy{
				{"הפניקס ביטוח", "2024-02-10", "2025-02-10", "ביטוח חיים", "55667788"},
				{"איילון ביטוח", "2024-05-01", "2025-05-01", "ביטוח נסיעות לחו\"ל", "99887766"},
			},
		},
	}

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("hash password: %v", err)
		}
		res, err := conn.Exec(`INSERT INTO users (name, email, password, phone) VALUES (?, ?, ?, ?)`,
			u.name, u.email, string(hash), u.phone)
		if err != nil {
			log.Fatalf("insert user: %v", err)
		}
		userID, _ := res.LastInsertId()
		for _, p := range u.policies {
			if _, err := conn.Exec(
				`INSERT INTO policies (user_id, insurance_company, start_date, end_date, coverage, kp_card) VALUES (?, ?, ?, ?, ?, ?)`,
				userID, p.company, p.start, p.end, p.coverage, p.kpCard,
			); err != nil {
				log.Fatalf("insert policy: %v", err)
			}
		}
	}
	log.Println("seeded demo users and policies (all demo passwords: password123)")
}
