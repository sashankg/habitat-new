package oauthserver

import (
	"time"

	"github.com/eagraf/habitat-new/internal/auth"
	"github.com/ory/fosite"
)

type authSession struct {
	Subject     string              `cbor:"1,keyasint"`
	ExpiresAt   time.Time           `cbor:"2,keyasint"`
	DpopKey     []byte              `cbor:"3,keyasint"`
	TokenInfo   *auth.TokenResponse `cbor:"4,keyasint"`
	ClientID    string              `cbor:"5,keyasint"`
	RedirectURI string              `cbor:"6,keyasint"`
	Scopes      []string            `cbor:"7,keyasint"`
	State       string              `cbor:"8,keyasint"`
	RequestedAt time.Time           `cbor:"9,keyasint"`
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
