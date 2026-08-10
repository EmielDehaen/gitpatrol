package models

import "time"

type Incident struct {
	ID        int       `json:"id"`
	RepoID    *int      `json:"repo_id"`
	RepoName  string    `json:"repo_name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Resolved  int       `json:"resolved"`
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
