package auth

import (
	"github.com/google/uuid"
)

type verifySessionParams struct {
	sessionID uuid.UUID
	userAgent string
	clientIp  string
}

type SessionManager interface {
	CreateSession() (uuid.UUID, error)
	VerifySession(arg verifySessionParams) (bool, error)
}
