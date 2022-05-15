package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func createRandomSession(t *testing.T) Session {
	// Arrange
	user := createRandomUser(t)
	arg := CreateSessionParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		UserAgent: "MacOS",
		ClientIp:  util.RandomIPAddress(),
		ExpiresAt: time.Now().Add(12 * time.Hour),
	}

	// Act
	session, err := testQueries.CreateSession(context.Background(), arg)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, session)

	require.Equal(t, arg.ID, session.ID)
	require.Equal(t, arg.UserID, session.UserID)
	require.Equal(t, arg.UserAgent, session.UserAgent)
	require.Equal(t, arg.ClientIp, session.ClientIp)
	// UTCか+09TZかで違うっぽい
	require.True(t, arg.ExpiresAt.Equal(session.ExpiresAt))

	require.NotZero(t, session.CreatedAt)

	return session
}

func TestCreateSession(t *testing.T) {
	createRandomSession(t)
}

func TestGetSession(t *testing.T) {
	// Arrange
	s := createRandomSession(t)

	// Act
	session, err := testQueries.GetSession(context.Background(), s.ID)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, session)

	require.Equal(t, s.ID, session.ID)
	require.Equal(t, s.UserID, session.UserID)
	require.Equal(t, s.UserAgent, session.UserAgent)
	require.Equal(t, s.ClientIp, session.ClientIp)
	// UTCか+09TZかで違うっぽい
	require.True(t, s.ExpiresAt.Equal(session.ExpiresAt))
	require.Equal(t, s.CreatedAt, session.CreatedAt)
}

func TestGetSessionWithWrongID(t *testing.T) {
	// Arrange

	// Act
	session, err := testQueries.GetSession(context.Background(), uuid.New())

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Empty(t, session)
}

func TestUpdateSession(t *testing.T) {
	// Arrange
	s := createRandomSession(t)
	newExpiresAt := time.Now().Add(30 * time.Minute)
	arg := UpdateSessionParams{
		ExpiresAt: newExpiresAt,
		ID:        s.ID,
	}

	// Act
	err := testQueries.UpdateSession(context.Background(), arg)

	// Assert
	require.NoError(t, err)

	ss, err := testQueries.GetSession(context.Background(), s.ID)
	require.NotEmpty(t, ss)
	require.NoError(t, err)

	require.Equal(t, s.ID, ss.ID)
	require.Equal(t, s.UserID, ss.UserID)
	require.Equal(t, s.UserAgent, ss.UserAgent)
	require.Equal(t, s.ClientIp, ss.ClientIp)
	// Updateさせたのでここが異なってほしい。
	require.False(t, s.ExpiresAt.Equal(ss.ExpiresAt))
	require.Equal(t, s.CreatedAt, ss.CreatedAt)
}

func TestDeleteSession(t *testing.T) {
	// Arrange
	s := createRandomSession(t)

	// Act
	err := testQueries.DeleteSession(context.Background(), s.ID)

	// Assert
	require.NoError(t, err)

	ss, err := testQueries.GetSession(context.Background(), s.ID)
	require.NotEmpty(t, ss)
	require.NoError(t, err)

	require.Equal(t, s.ID, ss.ID)
	require.Equal(t, s.UserID, ss.UserID)
	require.Equal(t, s.UserAgent, ss.UserAgent)
	require.Equal(t, s.ClientIp, ss.ClientIp)
	// 論理削除されていることの確認。
	require.True(t, time.Now().After(ss.ExpiresAt))
	require.Equal(t, s.CreatedAt, ss.CreatedAt)
}
