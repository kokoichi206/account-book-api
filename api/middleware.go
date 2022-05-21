package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kokoichi206/account-book-api/auth"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
)

const (
	cookieName = "session"
)

func (server *Server) authMiddleware(m auth.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionString, err := c.Cookie(cookieName)
		// Cookieから値が取得できない場合。
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("cannot find cookie")))
			return
		}
		session, err := uuid.Parse(sessionString)
		// 取得したCookieが、uuidの形式になってない場合。
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("wrong cookie value found")))
			return
		}

		arg := auth.VerifySessionParams{
			SessionID: session,
			UserAgent: c.Request.UserAgent(),
			ClientIp:  c.ClientIP(),
		}

		verify, err := m.VerifySession(arg)
		// セッションが有効ではない場合。
		if err != nil || !verify {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("session was not verified")))
			return
		}

		// セッションが有効な場合。
		// DBに保存しているセッションを自動更新する。
		duration := server.config.SessionDuration
		updateArg := db.UpdateSessionParams{
			ExpiresAt: time.Now().Add(duration),
			ID:        session,
		}
		err = server.querier.UpdateSession(context.Background(), updateArg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		// Cookieに渡しているセッションを自動更新する。
		maxAge := int(duration.Seconds())
		domain := server.config.ServerAddress
		c.SetCookie(cookieName, session.String(), maxAge, "/", domain, true, true)

		c.Next()
	}
}
