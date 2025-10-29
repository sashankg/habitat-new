package oauthserver

import (
	"time"

	"github.com/eagraf/habitat-new/internal/auth"
	"github.com/ory/fosite"
)

type authCodeSession struct {
	Subject   string
	ExpiresAt time.Time
	DpopKey   []byte
	TokenInfo *auth.TokenResponse
}

var _ fosite.Session = (*authCodeSession)(nil)

func newAuthCodeSession(
	subject string,
	dpopKey []byte,
	tokenInfo *auth.TokenResponse,
) *authCodeSession {
	return &authCodeSession{
		Subject:   subject,
		DpopKey:   dpopKey,
		TokenInfo: tokenInfo,
	}
}

// Clone implements fosite.Session.
func (s *authCodeSession) Clone() fosite.Session {
	return s
}

// GetExpiresAt implements fosite.Session.
func (s *authCodeSession) GetExpiresAt(key fosite.TokenType) time.Time {
	return s.ExpiresAt
}

// GetSubject implements fosite.Session.
func (s *authCodeSession) GetSubject() string {
	return s.Subject
}

// GetUsername implements fosite.Session.
func (s *authCodeSession) GetUsername() string {
	return s.Subject
}

// SetExpiresAt implements fosite.Session.
func (s *authCodeSession) SetExpiresAt(key fosite.TokenType, exp time.Time) {
	s.ExpiresAt = exp
}
