FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY pkg/ ./pkg/
COPY services/order/ ./services/order/
WORKDIR /app/services/order
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server
EXPOSE 8083
CMD ["/server"]
