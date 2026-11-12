package models

import (
	"time"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
	Salt     string `json:"salt,omitempty"`
}

type LoginAttempt struct {
	Count        int       `json:"count"`
	LastAttempt  time.Time `json:"last_attempt"`
	BlockedUntil time.Time `json:"blocked_until,omitempty"`
}
