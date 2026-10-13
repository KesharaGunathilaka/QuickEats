FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY pkg/ ./pkg/
COPY services/identity/ ./services/identity/
WORKDIR /app/services/identity
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server
EXPOSE 8081
CMD ["/server"]
