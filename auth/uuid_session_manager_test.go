package auth

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	mockdb "github.com/kokoichi206/account-book-api/db/mock"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := mockdb.NewMockQuerier(ctrl)

	// Act
	m := NewManager(querier)
	t.Log(m)

	// Assert
	require.NotNil(t, m)
}

func TestCreateSessioin(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := mockdb.NewMockQuerier(ctrl)
	m := NewManager(querier)

	// Act
	u, err := m.CreateSession()
	t.Log(u)
	t.Log(err)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, u)
	// RFC 4122 に基づく形式であることを確認する。
	require.True(t, regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).Match([]byte(u.String())))
}

func TestCreateSessioinUnique(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := mockdb.NewMockQuerier(ctrl)
	m := NewManager(querier)

	// Act
	u, err := m.CreateSession()
	t.Log(u)
	t.Log(err)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, u)

	uu, err := m.CreateSession()
	require.NoError(t, err)
	require.NotNil(t, uu)
	t.Log(uu)

	require.NotEqual(t, u, uu)
}

func TestVerifySession(t *testing.T) {
	u := uuid.New()

	ip := util.RandomIPAddress()

	arg := VerifySessionParams{
		SessionID: u,
		UserAgent: "MacOS",
		ClientIp:  ip,
	}
	session := db.Session{
		ID:        arg.SessionID,
		UserID:    util.RandomID(),
		UserAgent: arg.UserAgent,
		ClientIp:  arg.ClientIp,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	testCases := []struct {
		name          string
		arg           VerifySessionParams
		buildStubs    func(querier *mockdb.MockQuerier)
		checkResponse func(t *testing.T, valid bool, err error)
	}{
		{
			name: "OK",
			arg:  arg,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(t *testing.T, valid bool, err error) {
				require.True(t, valid)
				require.NoError(t, err)
			},
		},
		{
			name: "DBErrorNoRows",
			arg: VerifySessionParams{
				SessionID: uuid.New(),
				UserAgent: "MacOS",
				ClientIp:  ip,
			},
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, valid bool, err error) {
				require.False(t, valid)
				require.Error(t, err)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "InvalidWithExpired",
			arg:  arg,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{
						ID:        arg.SessionID,
						UserID:    util.RandomID(),
						UserAgent: arg.UserAgent,
						ClientIp:  util.RandomIPAddress(),
						CreatedAt: time.Now(),
						ExpiresAt: time.Now().Add(30 * time.Minute),
					}, nil)
			},
			checkResponse: func(t *testing.T, valid bool, err error) {
				require.False(t, valid)
				require.NoError(t, err)
			},
		},
		{
			name: "InvalidWithWrongUserAgent",
			arg:  arg,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{
						ID:        arg.SessionID,
						UserID:    util.RandomID(),
						UserAgent: arg.UserAgent,
						ClientIp:  arg.ClientIp,
						CreatedAt: time.Now(),
						ExpiresAt: time.Now().Add(-10 * time.Second),
					}, nil)
			},
			checkResponse: func(t *testing.T, valid bool, err error) {
				require.False(t, valid)
				require.NoError(t, err)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := mockdb.NewMockQuerier(ctrl)
			tc.buildStubs(querier)

			m := NewManager(querier)

			// Act
			s, err := m.VerifySession(tc.arg)

			// Assert
			tc.checkResponse(t, s, err)
		})
	}
}
