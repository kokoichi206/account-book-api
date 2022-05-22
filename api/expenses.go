package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"go.uber.org/zap"
)

// 新規の支出作成用のRequestのpayload。
type createExpenseRequest struct {
	UserID     int64  `json:"user_id" binding:"required"`
	CategoryID int64  `json:"category_id" binding:"required"`
	Amount     int64  `json:"amount" binding:"required"`
	Comment    string `json:"comment"`
}

// 出力用のJSONを取得する。
func (request createExpenseRequest) MustJSONString() string {
	bytes, err := json.Marshal(request)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// 新規の支出作成用のResponseのpayload。
type createExpenseResponse struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	CategoryID    int64     `json:"category_id"`
	Amount        int64     `json:"amount"`
	FoodReceiptID int64     `json:"food_receipt_id"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

// 支出の作成のエンドポイント。
func (server *Server) createExpense(c *gin.Context) {
	var req createExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// MAYBE: これはDebugかInfoか。
	zap.S().Debug(req.MustJSONString())

	arg := db.CreateExpenseParams{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
	}
	if req.Comment != "" {
		arg.Comment.Valid = true
		arg.Comment.String = req.Comment
	}

	expense, err := server.querier.CreateExpense(c, arg)
	if err != nil {
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	rsp := createExpenseResponse{
		ID:            expense.ID,
		UserID:        expense.UserID,
		CategoryID:    expense.CategoryID,
		Amount:        expense.Amount,
		FoodReceiptID: expense.FoodReceiptID.Int64,
		Comment:       expense.Comment.String,
		CreatedAt:     expense.CreatedAt,
	}

	c.JSON(http.StatusCreated, rsp)
}

// 支出一覧取得用のRequestのpayload。
type getAllExpensesRequest struct {
	UserID int64 `form:"user_id" binding:"required"`
}

// 出力用のJSONを取得する。
func (request getAllExpensesRequest) MustJSONString() string {
	bytes, err := json.Marshal(request)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// 支出一覧取得用のResponseのpayload。
type getAllExpensesResponse struct {
	ListExpenseResponse []expenseResponse `json:"expenses"`
}
type expenseResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	CategoryID int64     `json:"category_id"`
	Amount     int64     `json:"amount"`
	StoreName  string    `json:"store_name"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

// 支出一覧取得用のエンドポイント。
func (server *Server) getAllExpenses(c *gin.Context) {
	var req getAllExpensesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// MAYBE: これはDebugかInfoか。
	zap.S().Debug(req.MustJSONString())

	listExpenses, err := server.querier.ListExpenses(c, req.UserID)
	if err != nil {
		err = fmt.Errorf("failed to ListExpenses: %w", err)
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	var rsp getAllExpensesResponse
	for _, expense := range listExpenses {
		if expense.StoreName == nil {
			expense.StoreName = ""
		}
		e := expenseResponse{
			ID:         expense.ID,
			UserID:     expense.UserID,
			CategoryID: expense.CategoryID,
			Amount:     expense.Amount,
			StoreName:  expense.StoreName.(string),
			Comment:    expense.Comment.String,
			CreatedAt:  expense.CreatedAt,
		}
		rsp.ListExpenseResponse = append(rsp.ListExpenseResponse, e)
	}

	c.JSON(http.StatusOK, rsp)
}
