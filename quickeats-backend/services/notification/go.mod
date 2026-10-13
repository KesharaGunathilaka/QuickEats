module github.com/quickeats/services/notification

go 1.22

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/websocket v1.5.1
	github.com/quickeats/pkg v0.0.0
	github.com/nats-io/nats.go v1.34.1
)

replace github.com/quickeats/pkg => ../../pkg
