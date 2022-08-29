FROM golang:1.18 AS builder

WORKDIR /extension

COPY go.mod go.sum /extension/
COPY pkg /extension/pkg
COPY internal /extension/internal
COPY cmd /extension/cmd


RUN go mod edit -replace=github.com/skroutz/aws-lambda-secrets/internal/smsecrets=./internal/smsecrets && \
    go mod edit -replace=github.com/skroutz/aws-lambda-secrets/pkg/extension=./pkg/extension && \
    go mod tidy && \
    go mod verify && \
    go build -v -o extension/fetch-secrets cmd/fetch-secrets/main.go && \
    go build -v -o extension/wrapper/load-secrets cmd/load-secrets/main.go

FROM golang:1.18 AS runtime

WORKDIR /extension

COPY --from=builder /extension/extension /extension

ENTRYPOINT ["ls -l extension"]
