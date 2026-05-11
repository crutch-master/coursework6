package model

type User struct {
	ID           uint64
	Login        string
	Name         string
	Description  string
	PasswordHash []byte
}
