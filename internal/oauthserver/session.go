package oauthserver

import (
	"time"

	"github.com/eagraf/habitat-new/internal/auth"
	"github.com/ory/fosite"
)

type authSession struct {
	Subject   string
	ExpiresAt time.Time
	DpopKey   []byte
	TokenInfo *auth.TokenResponse
}

var _ fosite.Session = (*authSession)(nil)

func newAuthSession(
	subject string,
	dpopKey []byte,
	tokenInfo *auth.TokenResponse,
) *authSession {
	return &authSession{
		Subject:   subject,
		DpopKey:   dpopKey,
		TokenInfo: tokenInfo,
	}
}

// Clone implements fosite.Session.
func (s *authSession) Clone() fosite.Session {
	return s
}

// GetExpiresAt implements fosite.Session.
func (s *authSession) GetExpiresAt(key fosite.TokenType) time.Time {
	return s.ExpiresAt
}

// GetSubject implements fosite.Session.
func (s *authSession) GetSubject() string {
	return s.Subject
}

// GetUsername implements fosite.Session.
func (s *authSession) GetUsername() string {
	return s.Subject
}

// SetExpiresAt implements fosite.Session.
func (s *authSession) SetExpiresAt(key fosite.TokenType, exp time.Time) {
	s.ExpiresAt = exp
}
