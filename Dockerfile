FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY src/go.mod src/go.sum /app/
RUN go mod download

COPY src/*.go /app/

RUN go mod verify && \
	go build -v -o lambda-secrets

FROM alpine:3.16 AS runtime

WORKDIR /app

COPY --from=builder /app/lambda-secrets lambda-secrets 

ENTRYPOINT ["/app/lambda-secrets"]
