package auth

import (
	"context"
	"errors"
	"time"
)

type contextKey struct{}

var userContextKey = contextKey{}

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
)

type User struct {
	ID            string
	SessionID     string
	Email         string
	Name          string
	ActiveOrgID   string
	ActiveOrgRole OrgRole
	IssuedAt      time.Time
	ExpiresAt     time.Time
}

var (
	ErrNoUser       = errors.New("auth: no user in context")
	ErrUserExpired  = errors.New("auth: token expired")
	ErrInvalidToken = errors.New("auth: invalid token")
)

func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func FromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok || user == nil {
		return nil, ErrNoUser
	}
	if !user.ExpiresAt.IsZero() && time.Now().After(user.ExpiresAt) {
		return nil, ErrUserExpired
	}
	return user, nil
}

func MustFromContext(ctx context.Context) *User {
	user, err := FromContext(ctx)
	if err != nil {
		panic("auth.MustFromContext: " + err.Error())
	}
	return user
}
