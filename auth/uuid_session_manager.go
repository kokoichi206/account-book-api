package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
)

// セッション管理に関する構造体。
type UuidSessionManager struct {
	querier db.Querier
}

// セッション管理の構造体を作成し、返り値として受け取る。
func NewManager(querier db.Querier) SessionManager {

	manager := &UuidSessionManager{
		querier: querier,
	}

	return manager
}

// uuid を生成する。
// github.com/google/uuid を使用。
func (m *UuidSessionManager) CreateSession() (uuid.UUID, error) {
	return uuid.NewRandom()
}

// セッションが有効か確かめる。
//
// 以下の条件を全て満たす時、有効とする。
// * DBにセッションIDが存在する。
// * 有効期限が現在よりも長い。
// * アクセス元のUserAgentが発行時と同一。
// * アクセス元のClientIPが発行時と同一。
func (m *UuidSessionManager) VerifySession(arg VerifySessionParams) (bool, error) {
	s, err := m.querier.GetSession(context.Background(), arg.SessionID)
	if err != nil {
		// DBにセッションIDが存在しない時はエラーが返される。
		// see: TestGetSessionWithWrongID in db/sqlc
		return false, err
	}

	valid := s.ExpiresAt.After(time.Now()) &&
		arg.UserAgent == s.UserAgent &&
		arg.ClientIp == s.ClientIp
	return valid, nil
}
