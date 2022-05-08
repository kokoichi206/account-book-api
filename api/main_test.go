package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// ginのログを出力させない。
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
