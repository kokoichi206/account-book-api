package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func createRandomFoodReceipt(t *testing.T) FoodReceipt {
	// Arrange
	storeName := util.RandomStoreName()

	// Act
	foodReceipt, err := testQueries.CreateFoodReceipt(context.Background(), storeName)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, foodReceipt)

	require.NotZero(t, foodReceipt.ID)
	require.Equal(t, storeName, foodReceipt.StoreName)

	return foodReceipt
}

func TestCreateFoodReceipt(t *testing.T) {
	createRandomFoodReceipt(t)
}

func TestGetFoodReceipt(t *testing.T) {
	// Arrange
	foodReceipt1 := createRandomFoodReceipt(t)

	// Act
	foodReceipt2, err := testQueries.GetFoodReceipt(context.Background(), foodReceipt1.ID)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, foodReceipt2)

	require.Equal(t, foodReceipt1.ID, foodReceipt2.ID)
	require.Equal(t, foodReceipt1.StoreName, foodReceipt2.StoreName)
}

func TestGetFoodReceiptWithEmpty(t *testing.T) {
	// Arrange

	// Act
	foodReceipt, err := testQueries.GetFoodReceipt(context.Background(), util.RandomInt(0, 1000_0000))

	// Assert
	require.ErrorIs(t, err, sql.ErrNoRows)
	// 意外にも nil ではないことに注意する。
	require.NotNil(t, foodReceipt)

	require.Equal(t, foodReceipt.ID, int64(0))
	require.Equal(t, foodReceipt.StoreName, "")
}

func createRandomFoodContent(t *testing.T) FoodContent {
	// Arrange
	arg := CreateFoodContentParams{
		Name:         util.RandomFoodName(),
		Calories:     util.RandomCalories(),
		Lipid:        util.RandomNutrient(),
		Carbohydrate: util.RandomNutrient(),
		Protein:      util.RandomNutrient(),
	}

	// Act
	foodContent, err := testQueries.CreateFoodContent(context.Background(), arg)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, foodContent)

	require.NotZero(t, foodContent.ID)
	require.Equal(t, arg.Name, foodContent.Name)
	require.Equal(t, arg.Calories, foodContent.Calories)
	require.Equal(t, arg.Lipid, foodContent.Lipid)
	require.Equal(t, arg.Carbohydrate, foodContent.Carbohydrate)
	require.Equal(t, arg.Protein, foodContent.Protein)

	return foodContent
}

func TestCreateFoodContent(t *testing.T) {
	createRandomFoodContent(t)
}

func TestGetFoodContent(t *testing.T) {
	// Arrange
	foodContent1 := createRandomFoodContent(t)

	// Act
	foodContent2, err := testQueries.GetFoodContent(context.Background(), foodContent1.ID)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, foodContent2)

	require.Equal(t, foodContent1.ID, foodContent2.ID)
	require.Equal(t, foodContent1.Name, foodContent2.Name)
	require.Equal(t, foodContent1.Calories, foodContent2.Calories)
	require.Equal(t, foodContent1.Lipid, foodContent2.Lipid)
	require.Equal(t, foodContent1.Carbohydrate, foodContent2.Carbohydrate)
	require.Equal(t, foodContent1.Protein, foodContent2.Protein)
}

func TestGetFoodContentWithEmpty(t *testing.T) {
	// Arrange

	// Act
	foodContent, err := testQueries.GetFoodContent(context.Background(), util.RandomInt(0, 1000_0000))

	// Assert
	require.ErrorIs(t, err, sql.ErrNoRows)
	// 意外にも nil ではないことに注意する。
	require.NotNil(t, foodContent)

	require.Equal(t, foodContent.ID, int64(0))
	require.Equal(t, foodContent.Name, "")
	require.Equal(t, foodContent.Calories, float32(0))
	require.Equal(t, foodContent.Lipid, float32(0))
	require.Equal(t, foodContent.Carbohydrate, float32(0))
	require.Equal(t, foodContent.Protein, float32(0))
}

func createRandomFoodReceiptContent(t *testing.T, foodReceipt FoodReceipt, foodContent FoodContent) FoodReceiptContent {
	// Arrange

	arg := CreateFoodReceiptContentParams{
		FoodReceiptID: foodReceipt.ID,
		FoodContentID: foodContent.ID,
		Amount:        util.RandomAmount(),
	}

	// Act
	foodReceiptContent, err := testQueries.CreateFoodReceiptContent(context.Background(), arg)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, foodContent)

	require.NotZero(t, foodContent.ID)
	require.Equal(t, arg.FoodReceiptID, foodReceiptContent.FoodReceiptID)
	require.Equal(t, arg.FoodContentID, foodReceiptContent.FoodContentID)

	return foodReceiptContent
}

func TestCreateRandomFoodReceiptContent(t *testing.T) {
	foodReceipt := createRandomFoodReceipt(t)
	foodContent := createRandomFoodContent(t)
	createRandomFoodReceiptContent(t, foodReceipt, foodContent)
}

func TestListFoodReceiptContents(t *testing.T) {
	// Arrange
	foodReceipt := createRandomFoodReceipt(t)
	// 特定のレシートに対し、食品は複数存在しうる
	foodContent1 := createRandomFoodContent(t)
	foodContent2 := createRandomFoodContent(t)
	foodContent3 := createRandomFoodContent(t)
	expectedFoodContents := []FoodContent{
		foodContent1,
		foodContent2,
		foodContent3,
	}
	a := createRandomFoodReceiptContent(t, foodReceipt, foodContent1)
	b := createRandomFoodReceiptContent(t, foodReceipt, foodContent2)
	c := createRandomFoodReceiptContent(t, foodReceipt, foodContent3)
	expectedFoodReceiptContents := []FoodReceiptContent{
		a,
		b,
		c,
	}

	// Act
	foodReceiptContents, err := testQueries.ListFoodReceiptContents(context.Background(), foodReceipt.ID)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, foodReceiptContents)

	require.Equal(t, 3, len(foodReceiptContents))
	for i, foodReceiptContent := range foodReceiptContents {
		require.Equal(t, foodReceipt.ID, foodReceiptContent.FoodReceiptID)
		expectedFoodReceiptContent := expectedFoodReceiptContents[i]
		require.Equal(t, expectedFoodReceiptContent.Amount, foodReceiptContent.Amount)
		expectedFoodContent := expectedFoodContents[i]
		require.Equal(t, expectedFoodContent.ID, foodReceiptContent.FoodContentID)
		require.Equal(t, expectedFoodContent.Calories, foodReceiptContent.Calories)
		require.Equal(t, expectedFoodContent.Lipid, foodReceiptContent.Lipid)
		require.Equal(t, expectedFoodContent.Carbohydrate, foodReceiptContent.Carbohydrate)
		require.Equal(t, expectedFoodContent.Protein, foodReceiptContent.Protein)
	}
}

func TestListFoodReceiptContentsWithEmpty(t *testing.T) {
	// Arrange

	// Act
	foodReceiptContents, err := testQueries.ListFoodReceiptContents(context.Background(), util.RandomInt(0, 1000_0000))

	// Assert
	require.NoError(t, err)
	require.Empty(t, foodReceiptContents)
}
