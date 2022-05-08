// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0

package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Expense struct {
	ID         int64 `json:"id"`
	UserID     int64 `json:"user_id"`
	CategoryID int64 `json:"category_id"`
	// can be negative or positive
	Amount        int64          `json:"amount"`
	FoodReceiptID sql.NullInt64  `json:"food_receipt_id"`
	Comment       sql.NullString `json:"comment"`
	CreatedAt     time.Time      `json:"created_at"`
}

type FoodContent struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Calories     float32 `json:"calories"`
	Lipid        float32 `json:"lipid"`
	Carbohydrate float32 `json:"carbohydrate"`
	Protein      float32 `json:"Protein"`
}

type FoodReceipt struct {
	ID        int64  `json:"id"`
	StoreName string `json:"store_name"`
}

type FoodReceiptContent struct {
	ID            int64 `json:"id"`
	FoodReceiptID int64 `json:"food_receipt_id"`
	FoodContentID int64 `json:"food_content_id"`
	// must be positive
	Amount int64 `json:"amount"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Transfer struct {
	ID         int64 `json:"id"`
	FromUserID int64 `json:"from_user_id"`
	ToUserID   int64 `json:"to_user_id"`
	// must be positive
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Password          string    `json:"password"`
	Email             string    `json:"email"`
	Age               int32     `json:"age"`
	Balance           int64     `json:"balance"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}