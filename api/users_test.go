package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kokoichi206/account-book-api/auth"
	mockdb "github.com/kokoichi206/account-book-api/db/mock"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func addAuthMock(manager auth.MockUuidSessionManager, querier *mockdb.MockQuerier, userID int64) {
	uuid := uuid.New()
	manager.Uuid = uuid
	manager.CreateUUIDError = nil

	querier.EXPECT().
		CreateSession(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.Session{
			ID:        uuid,
			UserID:    userID,
			UserAgent: "MacOS",
			ClientIp:  util.RandomIPAddress(),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}, nil)
}

func TestCreateUser(t *testing.T) {

	name := util.RandomUserName()
	password := util.RandomPassword()
	email := util.RandomEmail()
	age := util.RandomAge()
	balance := util.RandomBalance()

	correctBody := gin.H{
		"username": name,
		"password": password,
		"email":    email,
		"age":      age,
		"balance":  balance,
	}
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	correctUser := db.User{
		ID:                1,
		Name:              name,
		Password:          hashedPassword,
		Email:             email,
		Age:               age,
		Balance:           balance,
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)

				addAuthMock(*manager, querier, correctUser.ID)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				checkResponseBody(t, recorder.Body, correctUser)
			},
		},
		{
			name: "BindRequestErrorWithMissingParam",
			body: gin.H{
				"username": name,
				"password": password,
				"age":      age,
				"balance":  balance,
			},
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BindRequestErrorWithValidationError",
			body: gin.H{
				"username": name,
				"password": password,
				"email":    util.RandomString(7),
				"age":      age,
				"balance":  balance,
			},
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "AlreadyRegisteredEmailError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, nil)

				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DBErrorWhenGetUser",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)

				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DBErrorWhenCreateUser",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "SessionManagerError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)

				manager.CreateUUIDError = errors.New("session manager error")

				querier.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DBErrorWhenCreateSession",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)

				uuid := uuid.New()
				manager.Uuid = uuid
				manager.CreateUUIDError = nil

				querier.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, errors.New("session manager error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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
			manager := auth.NewMockManager(querier)
			tc.buildStubs(querier, manager)

			server := NewServer(util.Config{}, querier, manager, util.InitLogger())
			recorder := httptest.NewRecorder()
			url := "/users"

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			// Act
			server.router.ServeHTTP(recorder, request)

			// Assert
			tc.checkResponse(t, recorder)
		})
	}
}

func TestLoginUser(t *testing.T) {

	password := util.RandomPassword()
	email := util.RandomEmail()

	// ログイン用のpayloadと同じものを用意する。
	correctBody := gin.H{
		"password": password,
		"email":    email,
	}
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	correctUser := db.User{
		ID:                1,
		Name:              util.RandomUserName(),
		Password:          hashedPassword,
		Email:             email,
		Age:               util.RandomAge(),
		Balance:           util.RandomBalance(),
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)
				addAuthMock(*manager, querier, correctUser.ID)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkResponseBody(t, recorder.Body, correctUser)
			},
		},
		{
			name: "BindRequestErrorWithMissingParam",
			body: gin.H{
				"password": password,
			},
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DBErrorWithNotRegistered",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DBError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "ErrorWithWrongPassword",
			body: gin.H{
				"password": "wrong_password",
				"email":    email,
			},
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "SessionManagerError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)

				manager.CreateUUIDError = errors.New("session manager error")

				querier.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "CreateSessionDBError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)

				uuid := uuid.New()
				manager.Uuid = uuid
				manager.CreateUUIDError = nil

				querier.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, errors.New("session manager error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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
			manager := auth.NewMockManager(querier)
			tc.buildStubs(querier, manager)

			server := NewServer(util.Config{}, querier, manager, util.InitLogger())
			recorder := httptest.NewRecorder()
			url := "/login"

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			// Act
			server.router.ServeHTTP(recorder, request)

			// Assert
			tc.checkResponse(t, recorder)
		})
	}
}

func checkResponseBody(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser userResponse
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.NotNil(t, gotUser.Id)
	require.Equal(t, user.Name, gotUser.Name)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.Age, gotUser.Age)
	require.Equal(t, user.Balance, gotUser.Balance)
}

func TestLogoutUser(t *testing.T) {

	testCases := []struct {
		name          string
		buildStubs    func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, manager *auth.MockUuidSessionManager)
	}{
		{
			name: "OK",
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					DeleteSession(gomock.Any(), manager.Uuid).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, manager *auth.MockUuidSessionManager) {
				require.Equal(t, http.StatusOK, recorder.Code)
				// Cookieの削除の指示が行われていること。
				checkSetCookie(t, recorder, manager.Uuid)
			},
		},
		{
			name: "DBErrorWhenDeleteSession",
			buildStubs: func(querier *mockdb.MockQuerier, manager *auth.MockUuidSessionManager) {
				querier.EXPECT().
					DeleteSession(gomock.Any(), manager.Uuid).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, manager *auth.MockUuidSessionManager) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				// ResponseBodyにエラーメッセージが乗っていること。
				checkError(t, sql.ErrConnDone.Error(), recorder.Body)
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
			manager := auth.NewMockManager(querier)

			session := uuid.New()
			manager.Verify = true
			manager.VerifyError = nil
			manager.Uuid = session

			tc.buildStubs(querier, manager)

			server := NewServer(util.Config{}, querier, manager, util.InitLogger())
			recorder := httptest.NewRecorder()
			url := "/logout"

			request, err := http.NewRequest(http.MethodPost, url, nil)
			require.NoError(t, err)

			// set cookie
			addAuthorization(t, request, session.String())

			// Act
			server.router.ServeHTTP(recorder, request)

			// Assert
			tc.checkResponse(t, recorder, manager)
		})
	}
}

func checkError(t *testing.T, errString string, responseBody *bytes.Buffer) {
	data, err := ioutil.ReadAll(responseBody)
	require.NoError(t, err)

	type errorMsg struct {
		Error string `json:"error"`
	}

	var body errorMsg
	err = json.Unmarshal(data, &body)
	require.NoError(t, err)
	require.NotNil(t, body.Error)
	require.Equal(t, errString, body.Error)
}

// Cookieの削除指示が正しく行われているか確認。
func checkSetCookie(t *testing.T, recorder *httptest.ResponseRecorder, sessionId uuid.UUID) {
	// Set-CookieがHeaderにあるかどうか。
	// 削除指示なので『Maz-Age=0』が指定されていること。
	setCookieValue := fmt.Sprintf("session=%v; Path=/; Max-Age=0; HttpOnly; Secure", sessionId)
	require.Equal(t, setCookieValue, recorder.Header().Get("Set-Cookie"))
}
