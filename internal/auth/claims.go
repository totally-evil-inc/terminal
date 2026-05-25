package auth

import (
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

type tokenClaims struct {
	ID            string
	Email         string
	Name          string
	SessionID     string
	ActiveOrgID   string
	ActiveOrgRole string
}

func userFromToken(tok jwt.Token) (*User, error) {
	claims := tokenClaims{}

	_ = tok.Get("id", &claims.ID)
	if claims.ID == "" {
		if sub, ok := tok.Subject(); ok {
			claims.ID = sub
		}
	}
	if claims.ID == "" {
		return nil, fmt.Errorf("%w: missing user id", ErrInvalidToken)
	}

	_ = tok.Get("email", &claims.Email)
	_ = tok.Get("name", &claims.Name)
	_ = tok.Get("sid", &claims.SessionID)
	_ = tok.Get("org_id", &claims.ActiveOrgID)
	_ = tok.Get("org_role", &claims.ActiveOrgRole)

	expiresAt, ok := tok.Expiration()
	if !ok || expiresAt.IsZero() {
		return nil, fmt.Errorf("%w: missing exp", ErrInvalidToken)
	}
	if time.Now().After(expiresAt) {
		return nil, ErrUserExpired
	}

	issuedAt, _ := tok.IssuedAt()

	return &User{
		ID:            claims.ID,
		SessionID:     claims.SessionID,
		Email:         claims.Email,
		Name:          claims.Name,
		ActiveOrgID:   claims.ActiveOrgID,
		ActiveOrgRole: OrgRole(claims.ActiveOrgRole),
		IssuedAt:      issuedAt,
		ExpiresAt:     expiresAt,
	}, nil
}
