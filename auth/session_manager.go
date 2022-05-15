package auth

import (
	"github.com/google/uuid"
)

type VerifySessionParams struct {
	SessionID uuid.UUID
	UserAgent string
	ClientIp  string
}

type SessionManager interface {
	CreateSession() (uuid.UUID, error)
	VerifySession(arg VerifySessionParams) (bool, error)
}
