FROM golang:1.22.3 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o sso ./cmd/sso/main.go

FROM golang:1.22.3

WORKDIR /app

COPY --from=builder /app/sso /app/sso

COPY ./config ./config

ENTRYPOINT ["/app/sso"]
CMD ["--config-path=./config/prod.yaml"]
