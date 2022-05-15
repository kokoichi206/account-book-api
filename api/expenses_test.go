package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestCreateExpense(t *testing.T) {

	url := "/expenses"
	userId := util.RandomID()
	categoryId := util.RandomID()
	amount := util.RandomExpense()
	comment := "big mistake"
	correctBody := gin.H{
		"user_id":     userId,
		"category_id": categoryId,
		"amount":      amount,
		"comment":     comment,
	}
	missingBody := gin.H{
		"user_id": userId,
		"amount":  amount,
		"comment": comment,
	}
	expense := db.Expense{
		ID:            util.RandomID(),
		UserID:        userId,
		CategoryID:    categoryId,
		Amount:        amount,
		FoodReceiptID: sql.NullInt64{},
		Comment: sql.NullString{
			Valid:  true,
			String: comment,
		},
		CreatedAt: time.Now(),
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
					CreateExpense(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expense, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				assertBody(t, expense, recorder.Body)
			},
		},
		{
			name: "BindRequestErrorWithMissingParam",
			body: missingBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateExpense(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "CreateExpenseDBError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateExpense(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Expense{}, sql.ErrConnDone)
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

			server := NewServer(util.Config{}, querier, nil)
			recorder := httptest.NewRecorder()

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

func assertBody(t *testing.T, expense db.Expense, responseBody *bytes.Buffer) {
	data, err := ioutil.ReadAll(responseBody)
	require.NoError(t, err)

	var body createExpenseResponse
	err = json.Unmarshal(data, &body)
	require.NoError(t, err)
	require.NotZero(t, body.ID)
	require.Equal(t, expense.UserID, body.UserID)
	require.Equal(t, expense.CategoryID, body.CategoryID)
	require.Equal(t, expense.Amount, body.Amount)
	if expense.FoodReceiptID.Valid {
		require.Equal(t, expense.FoodReceiptID.Int64, body.FoodReceiptID)
	} else {
		require.Zero(t, body.FoodReceiptID)
	}
	if expense.Comment.Valid {
		require.Equal(t, expense.Comment.String, body.Comment)
	} else {
		require.Nil(t, body.Comment)
	}
	require.NotZero(t, body.CreatedAt)
}

func TestGetAllExpenses(t *testing.T) {

	userId := util.RandomID()
	categoryId := util.RandomID()
	amount := util.RandomExpense()
	comment := "big mistake"

	listExpense := []db.ListExpensesRow{
		{
			ID:         util.RandomID(),
			UserID:     userId,
			CategoryID: categoryId,
			Amount:     amount,
			Comment: sql.NullString{
				Valid:  true,
				String: comment,
			},
			CreatedAt: time.Now(),
		},
	}
	listExpenseWithoutComment := []db.ListExpensesRow{
		{
			ID:         util.RandomID(),
			UserID:     userId,
			CategoryID: categoryId,
			Amount:     amount,
			CreatedAt:  time.Now(),
		},
	}
	testCases := []struct {
		name          string
		url           string
		buildStubs    func(querier *mockdb.MockQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			url:  fmt.Sprintf("/expenses?user_id=%d", userId),
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					ListExpenses(gomock.Any(), gomock.Any()).
					Times(1).
					Return(listExpense, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				assertListBody(t, listExpense, recorder.Body)
			},
		},
		{
			name: "OKWithoutComment",
			url:  fmt.Sprintf("/expenses?user_id=%d", userId),
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					ListExpenses(gomock.Any(), gomock.Any()).
					Times(1).
					Return(listExpenseWithoutComment, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				assertListBody(t, listExpenseWithoutComment, recorder.Body)
			},
		},
		{
			name: "BindRequestErrorWithMissingParam",
			url:  fmt.Sprintf("/expenses"),
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					ListExpenses(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ListExpenseDBError",
			url:  fmt.Sprintf("/expenses?user_id=%d", userId),
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					ListExpenses(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListExpensesRow{}, sql.ErrConnDone)
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

			server := NewServer(util.Config{}, querier, nil)
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, tc.url, nil)
			require.NoError(t, err)

			// Act
			server.router.ServeHTTP(recorder, request)

			// Assert
			tc.checkResponse(t, recorder)
		})
	}
}

func assertListBody(t *testing.T, listExpense []db.ListExpensesRow, responseBody *bytes.Buffer) {
	data, err := ioutil.ReadAll(responseBody)
	require.NoError(t, err)

	var body getAllExpensesResponse
	err = json.Unmarshal(data, &body)
	require.NoError(t, err)
	require.NotZero(t, body.ListExpenseResponse)

	require.Equal(t, 1, len(body.ListExpenseResponse))

	expense := body.ListExpenseResponse[0]
	expected := listExpense[0]
	require.Equal(t, expected.UserID, expense.UserID)
	require.Equal(t, expected.CategoryID, expense.CategoryID)
	require.Equal(t, expected.Amount, expense.Amount)
	if expected.StoreName == nil {
		require.Empty(t, expense.StoreName)
	} else {
		require.Equal(t, expected.StoreName, expense.StoreName)
	}
	if expected.Comment.Valid {
		require.Equal(t, expected.Comment.String, expense.Comment)
	} else {
		require.Empty(t, expense.Comment)
	}
	require.NotZero(t, expense.CreatedAt)
}
