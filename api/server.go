package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kokoichi206/account-book-api/auth"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"go.uber.org/zap"
)

// サーバーに関する情報を保持する構造体。
type Server struct {
	config         util.Config
	querier        db.Querier
	router         *gin.Engine
	sessionManager auth.SessionManager
	logger         *zap.Logger
}

// サーバーを作成し、返り値として受け取る。
func NewServer(config util.Config, querier db.Querier, manager auth.SessionManager, logger *zap.Logger) *Server {

	server := &Server{
		config:         config,
		querier:        querier,
		sessionManager: manager,
		logger:         logger,
	}

	server.setupRouter()
	return server
}

// ルーティングの設定を行い、構造体の変数に設定する。
func (server *Server) setupRouter() {
	router := gin.New()
	router.Use(util.GinLogger(server.logger), util.GinRecovery(server.logger, true))

	router.POST("/users", server.createUser)
	router.POST("/login", server.loginUser)
	router.POST("/logout", server.logout)

	authRoutes := router.Group("/").Use(server.authMiddleware(server.sessionManager))

	authRoutes.POST("/receipts", server.createReceipt)
	authRoutes.GET("/expenses", server.getAllExpenses)
	authRoutes.POST("/expenses", server.createExpense)

	server.router = router
}

// 指定したアドレスに対してHTTP serverを起動する。
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// エラー情報をJSONとして返すための関数。
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
