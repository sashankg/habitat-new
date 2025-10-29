package oauthserver

import (
	"crypto/ecdsa"
	"time"

	"github.com/eagraf/habitat-new/internal/auth"
	"github.com/ory/fosite"
)

type authCodeSession struct {
	ExpiresAt    time.Time
	DpopKey      *ecdsa.PrivateKey
	AccessToken  string
	RefreshToken string
	Subject      string
}

var _ fosite.Session = (*authCodeSession)(nil)

func newAuthCodeSession(
	subject string,
	dpopKey *ecdsa.PrivateKey,
	tokenInfo *auth.TokenResponse,
) *authCodeSession {
	return &authCodeSession{
		Subject:      subject,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(tokenInfo.ExpiresIn)),
		DpopKey:      dpopKey,
		AccessToken:  tokenInfo.AccessToken,
		RefreshToken: tokenInfo.RefreshToken,
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
