package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"
)

func Middleware(verifier Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, err := extractBearer(r)
			if err != nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			user, err := verifier.Verify(r.Context(), raw)
			if err != nil {
				log.Printf("auth verification failed: %v", err)
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
		})
	}
}

func extractBearer(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", ErrNoUser
	}

	scheme, token, ok := strings.Cut(auth, " ")
	if !ok || !strings.EqualFold(scheme, "bearer") {
		return "", ErrInvalidToken
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrInvalidToken
	}

	return token, nil
}

func IsUnauthenticated(err error) bool {
	return errors.Is(err, ErrNoUser) ||
		errors.Is(err, ErrUserExpired) ||
		errors.Is(err, ErrInvalidToken)
}

/*
Future tests to add when the testing layer is in scope:

- Bearer extraction rejects missing Authorization headers.
- Bearer extraction rejects non-Bearer schemes.
- Bearer extraction rejects empty Bearer tokens.
- Bearer extraction accepts a well-formed Bearer token.
- Context helpers return an error when no user is present.
- Context helpers return the stored user when one is present.
- Context helpers reject an expired stored user.
- The verifier accepts a valid signed token from a local JWKS server.
- The verifier rejects tokens with an invalid signature.
- The verifier rejects tokens with the wrong issuer.
- The verifier rejects tokens with the wrong audience.
- The verifier rejects expired tokens.
- Middleware allows a valid token through and injects the user into context.
- Middleware stops an invalid token request with 401 Unauthorized.
*/
