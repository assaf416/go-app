package db

import (
	"database/sql"

	"goapp/internal/models"
)

func FindUserByEmail(conn *sql.DB, email string) (*models.User, error) {
	row := conn.QueryRow(`SELECT id, name, email, password, phone FROM users WHERE email = ?`, email)
	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Phone); err != nil {
		return nil, err
	}
	return &u, nil
}

func FindUserByID(conn *sql.DB, id int64) (*models.User, error) {
	row := conn.QueryRow(`SELECT id, name, email, password, phone FROM users WHERE id = ?`, id)
	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Phone); err != nil {
		return nil, err
	}
	return &u, nil
}

func PoliciesForUser(conn *sql.DB, userID int64) ([]models.Policy, error) {
	rows, err := conn.Query(
		`SELECT id, user_id, insurance_company, start_date, end_date, coverage, kp_card
		 FROM policies WHERE user_id = ? ORDER BY start_date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []models.Policy
	for rows.Next() {
		var p models.Policy
		if err := rows.Scan(&p.ID, &p.UserID, &p.InsuranceCompany, &p.StartDate, &p.EndDate, &p.Coverage, &p.KPCard); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}
