package db

import (
	"context"
	"testing"

	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func createRandomCategory(t *testing.T) Category {
	// Arrange
	storeName := util.RandomString(8)

	// Act
	category, err := testQueries.CreateCategory(context.Background(), storeName)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, category)

	require.NotZero(t, category.ID)
	require.Equal(t, storeName, category.Name)

	return category
}

func TestCreateCategory(t *testing.T) {
	createRandomCategory(t)
}
