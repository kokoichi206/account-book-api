// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: users.sql

package db

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  name,
  password,
  email,
  age,
  balance
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING id, name, password, email, age, balance, password_changed_at, created_at
`

type CreateUserParams struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Age      int32  `json:"age"`
	Balance  int64  `json:"balance"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Name,
		arg.Password,
		arg.Email,
		arg.Age,
		arg.Balance,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Password,
		&i.Email,
		&i.Age,
		&i.Balance,
		&i.PasswordChangedAt,
		&i.CreatedAt,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT id, name, password, email, age, balance, password_changed_at, created_at FROM users
WHERE email = $1 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Password,
		&i.Email,
		&i.Age,
		&i.Balance,
		&i.PasswordChangedAt,
		&i.CreatedAt,
	)
	return i, err
}
