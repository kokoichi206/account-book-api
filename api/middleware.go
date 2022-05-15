package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kokoichi206/account-book-api/auth"
)

const (
	cookieName = "session"
)

func authMiddleware(m auth.SessionManager) gin.HandlerFunc {
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

		c.Next()
	}
}
