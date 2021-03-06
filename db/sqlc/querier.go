// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	CreateCategory(ctx context.Context, name string) (Category, error)
	CreateExpense(ctx context.Context, arg CreateExpenseParams) (Expense, error)
	CreateFoodContent(ctx context.Context, arg CreateFoodContentParams) (FoodContent, error)
	CreateFoodReceipt(ctx context.Context, storeName string) (FoodReceipt, error)
	CreateFoodReceiptContent(ctx context.Context, arg CreateFoodReceiptContentParams) (FoodReceiptContent, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	GetFoodContent(ctx context.Context, id int64) (FoodContent, error)
	GetFoodReceipt(ctx context.Context, id int64) (FoodReceipt, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	GetUser(ctx context.Context, email string) (User, error)
	ListExpenses(ctx context.Context, userID int64) ([]ListExpensesRow, error)
	ListFoodReceiptContents(ctx context.Context, foodReceiptID int64) ([]ListFoodReceiptContentsRow, error)
	UpdateSession(ctx context.Context, arg UpdateSessionParams) error
}

var _ Querier = (*Queries)(nil)
