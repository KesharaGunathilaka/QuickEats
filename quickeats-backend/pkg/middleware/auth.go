package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quickeats/pkg/httputil"
)

type contextKey string

const claimsKey contextKey = "user_claims"

// UserClaims holds the JWT claims extracted from the token.
type UserClaims struct {
	UserID         string `json:"sub"`
	OrganisationID string `json:"org"`
	Role           string `json:"role"`
	ShopID         string `json:"shop,omitempty"`
}

// Auth returns middleware that validates JWT tokens and injects claims into context.
func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httputil.RespondError(w, httputil.ErrUnauthenticated("missing authorization header"))
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				httputil.RespondError(w, httputil.ErrUnauthenticated("invalid authorization format"))
				return
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				httputil.RespondError(w, httputil.ErrUnauthenticated("invalid or expired token"))
				return
			}

			mapClaims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				httputil.RespondError(w, httputil.ErrUnauthenticated("invalid token claims"))
				return
			}

			claims := UserClaims{
				UserID:         getStr(mapClaims, "sub"),
				OrganisationID: getStr(mapClaims, "org"),
				Role:           getStr(mapClaims, "role"),
				ShopID:         getStr(mapClaims, "shop"),
			}

			ctx := context.WithValue(r.Context(), claimsKey, &claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves UserClaims from the request context.
func GetClaims(ctx context.Context) *UserClaims {
	claims, _ := ctx.Value(claimsKey).(*UserClaims)
	return claims
}

// RequireRole returns middleware that checks the user's role.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				httputil.RespondError(w, httputil.ErrUnauthenticated("no claims in context"))
				return
			}
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			httputil.RespondError(w, httputil.ErrForbidden("insufficient permissions"))
		})
	}
}

func getStr(m jwt.MapClaims, key string) string {
	v, _ := m[key].(string)
	return v
}
