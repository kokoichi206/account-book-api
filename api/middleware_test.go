package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kokoichi206/account-book-api/auth"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	session string,
) {
	cookie := fmt.Sprintf("session=%s", session)
	request.Header.Set("Cookie", cookie)
}

func TestAuthMiddleware(t *testing.T) {

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, manager auth.SessionManager)
		buildStubs    func(t *testing.T, manager auth.SessionManager)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				session := uuid.New()
				addAuthorization(t, request, session.String())
			},
			buildStubs: func(t *testing.T, manager auth.SessionManager) {
				mockManager := manager.(*auth.MockUuidSessionManager)
				mockManager.Verify = true
				mockManager.VerifyError = nil
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoCookie",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				// Not setup cookie
			},
			buildStubs: func(t *testing.T, manager auth.SessionManager) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				checkBodyContains(t, recorder, "cannot find cookie")
			},
		},
		{
			name: "WrongCookieValue",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				// Cookie value is not uuid type.
				addAuthorization(t, request, "wrong session type")
			},
			buildStubs: func(t *testing.T, manager auth.SessionManager) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				checkBodyContains(t, recorder, "wrong cookie value found")
			},
		},
		{
			name: "InvalidSession",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				session := uuid.New()
				addAuthorization(t, request, session.String())
			},
			buildStubs: func(t *testing.T, manager auth.SessionManager) {
				mockManager := manager.(*auth.MockUuidSessionManager)
				mockManager.Verify = false
				mockManager.VerifyError = nil
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				checkBodyContains(t, recorder, "session was not verified")
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := &db.Queries{}
			manager := auth.NewMockManager(querier)
			server := NewServer(util.Config{}, querier, manager)

			// managerについてはモックを使用する。
			tc.buildStubs(t, manager)
			// テスト用のパスを用意する。
			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.sessionManager),
				func(ctx *gin.Context) {
					// authが通ったときはStatusOKを返す。
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, manager)

			// Act
			server.router.ServeHTTP(recorder, request)

			// Assert
			tc.checkResponse(t, recorder)
		})
	}
}

// 対象とするレコーダーが指定文言を含むかチェックする。
func checkBodyContains(t *testing.T, recorder *httptest.ResponseRecorder, body string) {
	data, err := ioutil.ReadAll(recorder.Body)
	require.NoError(t, err)
	d := string(data)
	t.Log(d)
	t.Log(body)
	require.True(t, strings.Contains(d, body))
}
