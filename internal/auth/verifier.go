package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/muchirisworld/terminal/internal/config"
)

type Verifier interface {
	Verify(ctx context.Context, raw string) (*User, error)
}

type JWKSVerifier struct {
	jwksURL  string
	issuer   string
	audience string
	cache    *jwk.Cache
}

func NewJWKSVerifier(ctx context.Context, cfg *config.AuthConfig) (*JWKSVerifier, error) {
	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, fmt.Errorf("auth: create jwks cache: %w", err)
	}

	jwksURL := cfg.JWKSURL()
	if err := cache.Register(
		ctx,
		jwksURL,
		jwk.WithConstantInterval(1*time.Hour),
		jwk.WithMinInterval(15*time.Minute),
		jwk.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		jwk.WithWaitReady(false),
	); err != nil {
		return nil, fmt.Errorf("auth: register jwks url: %w", err)
	}

	primeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := cache.Refresh(primeCtx, jwksURL); err != nil {
		return nil, fmt.Errorf("auth: prime jwks cache: %w", err)
	}

	return &JWKSVerifier{
		jwksURL:  jwksURL,
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		cache:    cache,
	}, nil
}

func (v *JWKSVerifier) Verify(ctx context.Context, raw string) (*User, error) {
	if raw == "" {
		return nil, ErrInvalidToken
	}

	keyset, err := v.cache.Lookup(ctx, v.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("%w: load jwks: %v", ErrInvalidToken, err)
	}

	tok, err := jwt.Parse(
		[]byte(raw),
		jwt.WithKeySet(keyset),
		jwt.WithValidate(true),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithAcceptableSkew(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	return userFromToken(tok)
}
