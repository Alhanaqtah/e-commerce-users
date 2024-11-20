package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Birthdate time.Time `json:"birthdate"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	PassHash  []byte    `json:"-"`
	Version   int       `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
