FROM alpine:3.16 AS lambda-container

# Add the Lambda application to be run
WORKDIR /app/
COPY lambda-application /app/

# Add 'jq' to export JSON HTTP Response
RUN apk add jq

# == Setup 'lambda-secrets' ==
# Add 'lambda-secrets' binary from Container Image
COPY --from=lambda-secrets:latest /app/lambda-secrets /app/
COPY secrets-test.yaml /app/secrets.yaml

# Ensure 'lambda-secrets' runs BEFORE the Lambda application
ENTRYPOINT ["/app/lambda-secrets"]

# Pass the actual ENTRYPOINT to 'lambda-secrets':

# - Cleaner and has priority:
# CMD ["/app/lambda-application"]
# - Supports Shell notation such as pipes, loops
ENV ENTRYPOINT "/app/lambda-application"
