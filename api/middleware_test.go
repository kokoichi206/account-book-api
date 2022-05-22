package api

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kokoichi206/account-book-api/auth"
	mockdb "github.com/kokoichi206/account-book-api/db/mock"
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

func addCompleteAuth(t *testing.T, request *http.Request, manager *auth.MockUuidSessionManager) {
	session := uuid.New()
	addAuthorization(t, request, session.String())
	manager.Verify = true
	manager.VerifyError = nil
}

func TestAuthMiddleware(t *testing.T) {

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, manager auth.SessionManager)
		buildStubs    func(t *testing.T, querier *mockdb.MockQuerier, manager auth.SessionManager)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, server *Server)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				session := manager.(*auth.MockUuidSessionManager).Uuid
				addAuthorization(t, request, session.String())
			},
			buildStubs: func(t *testing.T, querier *mockdb.MockQuerier, manager auth.SessionManager) {
				querier.EXPECT().
					UpdateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)

				mockManager := manager.(*auth.MockUuidSessionManager)
				mockManager.Uuid = uuid.New()
				mockManager.Verify = true
				mockManager.VerifyError = nil
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, server *Server) {
				require.Equal(t, http.StatusOK, recorder.Code)

				mockManager := server.sessionManager.(*auth.MockUuidSessionManager)
				duration := server.config.SessionDuration.Seconds()
				setCookieValue := fmt.Sprintf("session=%v; Path=/; Max-Age=%v; HttpOnly; Secure", mockManager.Uuid, duration)
				require.Equal(t, setCookieValue, recorder.Header().Get("Set-Cookie"))
			},
		},
		{
			name: "NoCookie",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				// Not setup cookie
			},
			buildStubs: func(t *testing.T, querier *mockdb.MockQuerier, manager auth.SessionManager) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, server *Server) {
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
			buildStubs: func(t *testing.T, querier *mockdb.MockQuerier, manager auth.SessionManager) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, server *Server) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				checkBodyContains(t, recorder, "wrong cookie value found")
			},
		},
		{
			name: "InvalidSession",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				session := manager.(*auth.MockUuidSessionManager).Uuid
				addAuthorization(t, request, session.String())
			},
			buildStubs: func(t *testing.T, querier *mockdb.MockQuerier, manager auth.SessionManager) {
				mockManager := manager.(*auth.MockUuidSessionManager)
				mockManager.Uuid = uuid.New()
				mockManager.Verify = false
				mockManager.VerifyError = nil
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, server *Server) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				checkBodyContains(t, recorder, "session was not verified")
			},
		},
		{
			name: "DBErrorWhenUpdateSession",
			setupAuth: func(t *testing.T, request *http.Request, manager auth.SessionManager) {
				session := manager.(*auth.MockUuidSessionManager).Uuid
				addAuthorization(t, request, session.String())
			},
			buildStubs: func(t *testing.T, querier *mockdb.MockQuerier, manager auth.SessionManager) {
				querier.EXPECT().
					UpdateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)

				mockManager := manager.(*auth.MockUuidSessionManager)
				mockManager.Uuid = uuid.New()
				mockManager.Verify = true
				mockManager.VerifyError = nil
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, server *Server) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

				checkError(t, sql.ErrConnDone.Error(), recorder.Body)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config := util.Config{
				SessionDuration: 10 * time.Minute,
			}
			querier := mockdb.NewMockQuerier(ctrl)
			manager := auth.NewMockManager(querier)
			tc.buildStubs(t, querier, manager)

			server := NewServer(config, querier, manager, util.InitLogger())
			// テスト用のパスを用意する。
			authPath := "/auth"
			server.router.GET(
				authPath,
				server.authMiddleware(server.sessionManager),
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
			tc.checkResponse(t, recorder, server)
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
