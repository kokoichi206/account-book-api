package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func createRandomExpense(t *testing.T) Expense {
	// Arrange
	user := createRandomUser(t)
	category := createRandomCategory(t)
	arg := CreateExpenseParams{
		UserID:        user.ID,
		CategoryID:    category.ID,
		Amount:        util.RandomExpense(),
		FoodReceiptID: sql.NullInt64{},
		Comment:       sql.NullString{},
	}

	// Act
	expense, err := testQueries.CreateExpense(context.Background(), arg)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, expense)

	require.NotZero(t, expense.ID)
	require.Equal(t, arg.UserID, expense.UserID)
	require.Equal(t, arg.CategoryID, expense.CategoryID)
	require.Equal(t, arg.Amount, expense.Amount)
	require.Empty(t, expense.FoodReceiptID)
	require.Empty(t, expense.Comment)
	require.NotZero(t, expense.CreatedAt)

	return expense
}

func TestCreateExpense(t *testing.T) {
	createRandomExpense(t)
}

func TestCreateExpenseWithReceipt(t *testing.T) {
	// Arrange
	user := createRandomUser(t)
	t.Log(user)
	category := createRandomCategory(t)
	t.Log(category)
	receipt := createRandomFoodReceipt(t)
	t.Log(receipt)
	arg := CreateExpenseParams{
		UserID:     user.ID,
		CategoryID: category.ID,
		Amount:     util.RandomExpense(),
		Comment:    sql.NullString{},
	}
	// Nullableな変数の扱い方に慣れる。
	arg.FoodReceiptID.Valid = true
	arg.FoodReceiptID.Int64 = receipt.ID
	t.Log(arg)

	// Act
	expense, err := testQueries.CreateExpense(context.Background(), arg)
	t.Log(expense)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, expense)

	require.NotZero(t, expense.ID)
	require.Equal(t, arg.UserID, expense.UserID)
	require.Equal(t, arg.CategoryID, expense.CategoryID)
	require.Equal(t, arg.Amount, expense.Amount)
	require.NotEmpty(t, expense.FoodReceiptID)
	// Nullableな変数の扱い方に慣れる。
	require.True(t, expense.FoodReceiptID.Valid)
	require.Equal(t, receipt.ID, expense.FoodReceiptID.Int64)
	require.Empty(t, expense.Comment)
	require.NotZero(t, expense.CreatedAt)
}

func TestCreateExpenseWithWrongFKValue(t *testing.T) {
	// Arrange
	category := createRandomCategory(t)
	t.Log(category)
	receipt := createRandomFoodReceipt(t)
	t.Log(receipt)
	arg := CreateExpenseParams{
		UserID:        util.RandomID(),
		CategoryID:    category.ID,
		Amount:        util.RandomExpense(),
		FoodReceiptID: sql.NullInt64{},
		Comment:       sql.NullString{},
	}

	// Act
	expense, err := testQueries.CreateExpense(context.Background(), arg)
	t.Log(expense)
	t.Log(err)

	// Assert
	require.Error(t, err)
	require.Empty(t, expense)
}

func TestListExpenses(t *testing.T) {
	// Arrange
	user := createRandomUser(t)
	t.Log(user)
	category := createRandomCategory(t)
	t.Log(category)
	arg1 := CreateExpenseParams{
		UserID:        user.ID,
		CategoryID:    category.ID,
		Amount:        util.RandomExpense(),
		FoodReceiptID: sql.NullInt64{},
		Comment:       sql.NullString{},
	}
	arg2 := CreateExpenseParams{
		UserID:        user.ID,
		CategoryID:    category.ID,
		Amount:        util.RandomExpense(),
		FoodReceiptID: sql.NullInt64{},
		Comment:       sql.NullString{},
	}
	arg3 := CreateExpenseParams{
		UserID:        user.ID,
		CategoryID:    category.ID,
		Amount:        util.RandomExpense(),
		FoodReceiptID: sql.NullInt64{},
		Comment:       sql.NullString{},
	}
	args := []CreateExpenseParams{
		arg1,
		arg2,
		arg3,
	}
	for _, arg := range args {
		testQueries.CreateExpense(context.Background(), arg)
		// dummy data
		createRandomExpense(t)
	}

	// Act
	expenses, err := testQueries.ListExpenses(context.Background(), user.ID)
	t.Log(expenses)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, expenses)

	require.Equal(t, len(args), len(expenses))
	for i, expense := range expenses {
		arg := args[i]
		require.NotZero(t, expense.ID)
		require.Equal(t, user.ID, expense.UserID)
		require.Equal(t, category.ID, expense.CategoryID)
		require.Equal(t, arg.Amount, expense.Amount)
		// Without food_receipts_id returns Empty
		require.Empty(t, expense.StoreName)
		require.Empty(t, expense.Comment)
		require.NotZero(t, expense.CreatedAt)
	}
}
