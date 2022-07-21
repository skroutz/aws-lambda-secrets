FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY src/* /app/

RUN go mod verify && \
	go build -v ./...

FROM alpine:3.16 AS runtime

WORKDIR /app

COPY --from=builder /app/lambda-secrets lambda-secrets 

ENTRYPOINT /app/lambda-secrets
