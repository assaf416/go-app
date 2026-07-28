package models

type User struct {
	ID       int64
	Name     string
	Email    string
	Password string
	Phone    string
}

type Policy struct {
	ID                int64
	UserID            int64
	InsuranceCompany  string
	StartDate         string
	EndDate           string
	Coverage          string
	KPCard            string
}
