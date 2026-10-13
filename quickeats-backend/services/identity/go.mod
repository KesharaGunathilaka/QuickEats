module github.com/quickeats/services/identity

go 1.22

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/quickeats/pkg v0.0.0
	golang.org/x/crypto v0.22.0
)

replace github.com/quickeats/pkg => ../../pkg
