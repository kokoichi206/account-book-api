package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/kokoichi206/account-book-api/db/mock"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

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
		buildStubs    func(querier *mockdb.MockQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)
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
			buildStubs: func(querier *mockdb.MockQuerier) {
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
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DBError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
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
			tc.buildStubs(querier)

			server := NewServer(util.Config{}, querier)
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
		buildStubs    func(querier *mockdb.MockQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)
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
			buildStubs: func(querier *mockdb.MockQuerier) {
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
			buildStubs: func(querier *mockdb.MockQuerier) {
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
			buildStubs: func(querier *mockdb.MockQuerier) {
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
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(correctUser, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			server := NewServer(util.Config{}, querier)
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
