FROM golang:1.26-alpine AS builder
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.21
COPY --from=builder /go/bin/goose /bin/goose
COPY migrations /migrations
ENTRYPOINT ["goose", "-dir", "/migrations", "up"]
