# Use a build argument to switch between v1 and v2
ARG APP_VERSION=v1

FROM golang:1.21-alpine AS builder
ARG APP_VERSION
WORKDIR /app

# Copy the specific version directory based on the argument
COPY ${APP_VERSION}/app.go ./main.go

# Build a binary
# 1. Initialize the module
# 2. 'tidy' will find and download github.com/lib/pq
# 3. Build the binary
RUN go mod init canarydemo && \
    go mod tidy && \
    go build -o app main.go

FROM alpine:latest  
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]