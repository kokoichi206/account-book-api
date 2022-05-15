package auth

import (
	"github.com/google/uuid"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
)

type MockUuidSessionManager struct {
	Uuid            uuid.UUID
	CreateUUIDError error
	Verify          bool
	VerifyError     error
	VerifyArg       VerifySessionParams
}

func NewMockManager(querier db.Querier) *MockUuidSessionManager {

	manager := &MockUuidSessionManager{}
	return manager
}

func (m *MockUuidSessionManager) CreateSession() (uuid.UUID, error) {
	return m.Uuid, m.CreateUUIDError
}

func (m *MockUuidSessionManager) VerifySession(arg VerifySessionParams) (bool, error) {
	m.VerifyArg = arg
	return m.Verify, m.VerifyError
}
