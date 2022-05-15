package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
)

// 新規ユーザー作成用のpayload。
type createUserRequest struct {
	Name     string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Age      int32  `json:"age"`
	Balance  int64  `json:"balance"`
}

// 新規ユーザー作成用のpayload。
type userResponse struct {
	Id                int       `json:"id"`
	Name              string    `json:"username"`
	Email             string    `json:"email"`
	Age               int32     `json:"age"`
	Balance           int64     `json:"balance"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// ログイン用のpayload。
type loginUserRequest struct {
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

// 新規ユーザー作成のエンドポイント。
func (server *Server) createUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	arg := db.CreateUserParams{
		Name:     req.Name,
		Password: hashedPassword,
		Email:    req.Email,
		Age:      req.Age,
		Balance:  req.Balance,
	}

	user, err := server.querier.CreateUser(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// セッションを発行し、Cookieにセットする。
	id, err := server.sessionManager.CreateSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	sarg := db.CreateSessionParams{
		ID:        id,
		UserID:    user.ID,
		UserAgent: c.Request.UserAgent(),
		ClientIp:  c.ClientIP(),
		ExpiresAt: time.Now().Add(server.config.SessionDuration),
	}
	session, err := server.querier.CreateSession(context.Background(), sarg)
	if err != nil {
		// DBに何かしらの不備がある。
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	maxAge := int(server.config.SessionDuration.Seconds())
	domain := server.config.ServerAddress
	c.SetCookie(cookieName, session.ID.String(), maxAge, "/", domain, true, true)

	res := userResponse{
		Name:              user.Name,
		Email:             user.Email,
		Age:               user.Age,
		Balance:           user.Balance,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	c.JSON(http.StatusCreated, res)
}

// 既存ユーザー用のログインのエンドポイント。
func (server *Server) loginUser(c *gin.Context) {
	var req loginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Emailが登録されているかチェックする。
	user, err := server.querier.GetUser(c, req.Email)
	if err != nil {
		// 登録されていなければ、ユーザーのリクエストに不備がある。
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		// それ以外は、DBに何かしらの不備がある。
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// パスワードをチェックする。
	if err := util.CheckPassword(req.Password, user.Password); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// セッションを発行し、Cookieにセットする。
	id, err := server.sessionManager.CreateSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	sarg := db.CreateSessionParams{
		ID:        id,
		UserID:    user.ID,
		UserAgent: c.Request.UserAgent(),
		ClientIp:  c.ClientIP(),
		ExpiresAt: time.Now().Add(server.config.SessionDuration),
	}
	session, err := server.querier.CreateSession(context.Background(), sarg)
	if err != nil {
		// DBに何かしらの不備がある。
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	maxAge := int(server.config.SessionDuration.Seconds())
	domain := server.config.ServerAddress
	c.SetCookie(cookieName, session.ID.String(), maxAge, "/", domain, true, true)

	res := userResponse{
		Name:              user.Name,
		Email:             user.Email,
		Age:               user.Age,
		Balance:           user.Balance,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	c.JSON(http.StatusOK, res)
}
