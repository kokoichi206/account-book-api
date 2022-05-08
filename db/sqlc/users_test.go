package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	// Arrange
	arg := CreateUserParams{
		Name:     util.RandomUserName(),
		Password: "secret password", // TODO: ハッシュ化
		Email:    util.RandomEmail(),
		Age:      util.RandomAge(),
		Balance:  util.RandomBalance(),
	}

	// Act
	user, err := testQueries.CreateUser(context.Background(), arg)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Name, user.Name)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Age, user.Age)
	require.Equal(t, arg.Balance, user.Balance)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

// NOT NULLを指定している属性を指定しない場合のテスト。
func TestCreateUserWithEmptyParams(t *testing.T) {
	// Act
	// NOT NULLを指定しているパラメーターに対して、Emptyでも登録できてしまうことに注意したい。
	// 型によってデフォルトの値がエラーも返ってこない。
	user, err := testQueries.CreateUser(context.Background(), CreateUserParams{})

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, "", user.Name)
	require.Equal(t, "", user.Password)
	require.Equal(t, "", user.Email)
	require.Equal(t, int32(0), user.Age)
	require.Equal(t, int64(0), user.Balance)
}

// Uniqueを指定しているemailが被った場合のテスト。
func TestCreateUserWithDuplicateEmail(t *testing.T) {
	// Arrange
	user1 := createRandomUser(t)
	arg := CreateUserParams{
		Name:     util.RandomUserName(),
		Password: "secret password", // TODO: ハッシュ化
		Email:    user1.Email,       // 先ほど作成したユーザーと同じEmailを使う。
		Age:      util.RandomAge(),
		Balance:  util.RandomBalance(),
	}

	// Act
	user, err := testQueries.CreateUser(context.Background(), arg)

	require.Error(t, err)
	require.Empty(t, user)
}

func TestGetUser(t *testing.T) {
	// Arrange
	user1 := createRandomUser(t)

	// Act
	user2, err := testQueries.GetUser(context.Background(), user1.Email)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Name, user2.Name)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Age, user2.Age)
	require.Equal(t, user1.Balance, user2.Balance)

	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestGetUserWithNotExist(t *testing.T) {
	// Arrange

	// Act
	user2, err := testQueries.GetUser(context.Background(), "test@example.com")

	// Assert
	require.Error(t, err)
	require.Empty(t, user2)

	require.Equal(t, err, sql.ErrNoRows)
}
