package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
)

// サーバーに関する情報を保持する構造体。
type Server struct {
	config  util.Config
	querier db.Querier
	router  *gin.Engine
}

// サーバーを作成し、返り値として受け取る。
func NewServer(config util.Config, querier db.Querier) *Server {

	server := &Server{
		config:  config,
		querier: querier,
	}

	server.setupRouter()
	return server
}

// ルーティングの設定を行い、構造体の変数に設定する。
func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/login", server.loginUser)

	router.POST("/receipts", server.createReceipt)

	server.router = router
}

// 指定したアドレスに対してHTTP serverを起動する。
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
