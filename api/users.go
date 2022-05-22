package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"go.uber.org/zap"
)

// 新規ユーザー作成用のpayload。
type createUserRequest struct {
	Name     string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Age      int32  `json:"age"`
	Balance  int64  `json:"balance"`
}

// 出力用のJSONを取得する。
func (request createUserRequest) MustJSONString() string {
	bytes, err := json.Marshal(request)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// 新規ユーザー作成用のpayload。
type userResponse struct {
	Id                int64     `json:"id"`
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

// 出力用のJSONを取得する。
func (request loginUserRequest) MustJSONString() string {
	bytes, err := json.Marshal(request)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// 新規ユーザー作成のエンドポイント。
func (server *Server) createUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// MAYBE: これはDebugかInfoか。
	zap.S().Debug(req.MustJSONString())

	// Emailが登録されているかチェックする。
	_, err := server.querier.GetUser(c, req.Email)
	if err != sql.ErrNoRows {
		// エラーなし↔︎すでにEmailは登録済み
		if err == nil {
			message := fmt.Sprintf("The Email [%s] has already registered.", req.Email)
			zap.S().Warn(message)
			c.JSON(http.StatusBadRequest, errorResponse(errors.New("The Email has already registered.")))
			return
		}
		// それ以外は、DBに何かしらの不備がある。
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		// パスワードのハッシュ化ができず先に進まないのは致命的。
		zap.S().Error(err)

		c.Error(err)
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
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// セッションを発行し、Cookieにセットする。
	id, err := server.sessionManager.CreateSession()
	if err != nil {
		zap.S().Error(err)

		c.Error(err)
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
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	maxAge := int(server.config.SessionDuration.Seconds())
	domain := server.config.ServerAddress
	c.SetCookie(cookieName, session.ID.String(), maxAge, "/", domain, true, true)

	res := userResponse{
		Id:                user.ID,
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

	// MAYBE: これはDebugかInfoか。
	zap.S().Debug(req.MustJSONString())

	// Emailが登録されているかチェックする。
	user, err := server.querier.GetUser(c, req.Email)
	if err != nil {
		// 登録されていなければ、ユーザーのリクエストに不備がある。
		if err == sql.ErrNoRows {
			message := fmt.Sprintf("The Email [%s] has not registered yet.", req.Email)
			zap.S().Warn(message)
			c.JSON(http.StatusBadRequest, err)
			return
		}
		// それ以外は、DBに何かしらの不備がある。
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// パスワードをチェックする。
	if err := util.CheckPassword(req.Password, user.Password); err != nil {
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// セッションを発行し、Cookieにセットする。
	id, err := server.sessionManager.CreateSession()
	if err != nil {
		zap.S().Error(err)

		c.Error(err)
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
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	maxAge := int(server.config.SessionDuration.Seconds())
	domain := server.config.ServerAddress
	c.SetCookie(cookieName, session.ID.String(), maxAge, "/", domain, true, true)

	res := userResponse{
		Id:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		Age:               user.Age,
		Balance:           user.Balance,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	c.JSON(http.StatusOK, res)
}

// ログアウト用のエンドポイント。
func (server *Server) logout(c *gin.Context) {

	// authを通してるので、基本的に前半でこけることはない。
	sessionString, err := c.Cookie(cookieName)
	// Cookieから値が取得できない場合。
	if err != nil {
		message := fmt.Errorf("could not find cookie: %w", err)
		zap.S().Warn(message)
		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	sessionID, err := uuid.Parse(sessionString)
	// 取得したCookieが、uuidの形式になってない場合。
	if err != nil {
		message := fmt.Sprintf("could not convert session [%s] to uuid.", sessionID)
		zap.S().Warn(message)
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.querier.DeleteSession(context.Background(), sessionID)
	if err != nil {
		zap.S().Error(err)

		c.Error(err)
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	domain := server.config.ServerAddress
	// Cookieの有効期限を負の値にし、論理的に削除にする。
	c.SetCookie(cookieName, sessionID.String(), -1, "/", domain, true, true)

	c.Status(http.StatusOK)
}
