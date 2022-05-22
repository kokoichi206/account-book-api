package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"go.uber.org/zap"
)

// レシート登録用のpayload。
type createReceiptRequest struct {
	StoreName    string        `json:"store_name" binding:"required"`
	FoodContents []foodContent `json:"food_contents" binding:"required"`
	TotalPrice   int           `json:"total_price" binding:"required"`
}

// 出力用のJSONを取得する。
func (request createReceiptRequest) MustJSONString() string {
	bytes, err := json.Marshal(request)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// レシート登録時に使う、１つの食品用の構造体。
type foodContent struct {
	Name  string `json:"name" binding:"required"`
	Price int    `json:"price" binding:"required"`
}

// １枚のレシートを登録するエンドポイント。
func (server *Server) createReceipt(c *gin.Context) {
	var req createReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// MAYBE: これはDebugかInfoか。
	zap.S().Debug(req.MustJSONString())

	storeName := req.StoreName
	foodReceipt, err := server.querier.CreateFoodReceipt(c, storeName)
	if err != nil {
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	for _, content := range req.FoodContents {
		// TODO: 商品名（と店名）から栄養素を取得する。もし存在しなければ保存する（？）
		foodContent := getFoodContent(storeName, content)
		arg := db.CreateFoodReceiptContentParams{
			FoodReceiptID: foodReceipt.ID,
			FoodContentID: foodContent.ID,
			Amount:        1,
		}
		// MAYBE: 一部商品だけ登録されることがあっていいのか、考える。
		// もしかしたら transaction を行う必要があるかも。
		_, err := server.querier.CreateFoodReceiptContent(c, arg)
		if err != nil {
			zap.S().Error(err)

			c.Error(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	}
	c.Status(http.StatusOK)
}

// 商品名と店名から栄養素を取得する。
func getFoodContent(storeName string, foodContent foodContent) db.FoodContent {
	// TODO: 検索するシステムを作成する。
	return db.FoodContent{
		ID:           util.RandomInt(0, 10_0000),
		Name:         util.RandomFoodName(),
		Calories:     util.RandomCalories(),
		Lipid:        util.RandomNutrient(),
		Carbohydrate: util.RandomNutrient(),
		Protein:      util.RandomNutrient(),
	}
}
