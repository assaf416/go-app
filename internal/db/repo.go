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

func InsertProject(conn *sql.DB, name, githubURL string) (*models.Project, error) {
	res, err := conn.Exec(`INSERT INTO projects (project_name, github_url) VALUES (?, ?)`, name, githubURL)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Project{ID: id, ProjectName: name, GithubURL: githubURL}, nil
}

func ListProjects(conn *sql.DB) ([]models.Project, error) {
	rows, err := conn.Query(`SELECT id, project_name, github_url FROM projects ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.ProjectName, &p.GithubURL); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func FindProjectByID(conn *sql.DB, id int64) (*models.Project, error) {
	row := conn.QueryRow(`SELECT id, project_name, github_url FROM projects WHERE id = ?`, id)
	var p models.Project
	if err := row.Scan(&p.ID, &p.ProjectName, &p.GithubURL); err != nil {
		return nil, err
	}
	return &p, nil
}

func FindProjectByName(conn *sql.DB, name string) (*models.Project, error) {
	row := conn.QueryRow(`SELECT id, project_name, github_url FROM projects WHERE project_name = ?`, name)
	var p models.Project
	if err := row.Scan(&p.ID, &p.ProjectName, &p.GithubURL); err != nil {
		return nil, err
	}
	return &p, nil
}
