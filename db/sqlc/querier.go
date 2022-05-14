// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0

package db

import (
	"context"
)

type Querier interface {
	CreateCategory(ctx context.Context, name string) (Category, error)
	CreateExpense(ctx context.Context, arg CreateExpenseParams) (Expense, error)
	CreateFoodContent(ctx context.Context, arg CreateFoodContentParams) (FoodContent, error)
	CreateFoodReceipt(ctx context.Context, storeName string) (FoodReceipt, error)
	CreateFoodReceiptContent(ctx context.Context, arg CreateFoodReceiptContentParams) (FoodReceiptContent, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetFoodContent(ctx context.Context, id int64) (FoodContent, error)
	GetFoodReceipt(ctx context.Context, id int64) (FoodReceipt, error)
	GetUser(ctx context.Context, email string) (User, error)
	ListExpenses(ctx context.Context, userID int64) ([]ListExpensesRow, error)
	ListFoodReceiptContents(ctx context.Context, foodReceiptID int64) ([]ListFoodReceiptContentsRow, error)
}

var _ Querier = (*Queries)(nil)
