package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kokoichi206/account-book-api/auth"
	mockdb "github.com/kokoichi206/account-book-api/db/mock"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func TestCreateReceipt(t *testing.T) {
	storeName := util.RandomStoreName()
	foodContent1 := foodContent{
		Name:  util.RandomFoodName(),
		Price: int(util.RandomInt(10, 1100)),
	}
	foodContent2 := foodContent{
		Name:  util.RandomFoodName(),
		Price: int(util.RandomInt(10, 1100)),
	}
	foodContent3 := foodContent{
		Name:  util.RandomFoodName(),
		Price: int(util.RandomInt(10, 1100)),
	}
	foodContents := []foodContent{
		foodContent1,
		foodContent2,
		foodContent3,
	}
	// １枚のレシートとして送られてくる情報。
	correctBody := gin.H{
		"store_name":    storeName,
		"food_contents": foodContents,
		"total_price":   foodContent1.Price + foodContent2.Price + foodContent3.Price,
	}
	missingBody := gin.H{
		"store_name":    storeName,
		"food_contents": foodContents,
	}

	foodReceipt := db.FoodReceipt{
		ID:        1,
		StoreName: storeName,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, manager *auth.MockUuidSessionManager)
		buildStubs    func(querier *mockdb.MockQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateFoodReceipt(gomock.Any(), gomock.Any()).
					Times(1).
					Return(foodReceipt, nil)
				querier.EXPECT().
					CreateFoodReceiptContent(gomock.Any(), gomock.Any()).
					Times(3)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "BindRequestErrorWithMissingParam",
			body: missingBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateFoodReceipt(gomock.Any(), gomock.Any()).
					Times(0)
				querier.EXPECT().
					CreateFoodReceiptContent(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "CreateFoodReceiptDBError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateFoodReceipt(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.FoodReceipt{}, sql.ErrConnDone)
				querier.EXPECT().
					CreateFoodReceiptContent(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "CreateFoodReceiptContentDBError",
			body: correctBody,
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					CreateFoodReceipt(gomock.Any(), gomock.Any()).
					Times(1).
					Return(foodReceipt, nil)
				querier.EXPECT().
					CreateFoodReceiptContent(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.FoodReceiptContent{}, sql.ErrConnDone)
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
			manager := auth.NewMockManager(querier)

			server := NewServer(util.Config{}, querier, manager)
			recorder := httptest.NewRecorder()
			url := "/receipts"

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			addCompleteAuth(t, request, manager)

			// Act
			server.router.ServeHTTP(recorder, request)

			// Assert
			tc.checkResponse(t, recorder)
		})
	}
}
