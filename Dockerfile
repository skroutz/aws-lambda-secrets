FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY src/* .

RUN go mod verify && \
	go build -v ./...

FROM alpine:3.17 AS runtime

WORKDIR /app

COPY --from=builder /app/lambda-secrets lambda-secrets 

ENTRYPOINT /app/lambda-secrets
