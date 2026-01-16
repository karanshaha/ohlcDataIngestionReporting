FROM golang:1.22-alpine AS build

WORKDIR /app

# Copy EVERYTHING first (includes go.mod, source, etc.)
COPY . .
RUN go mod download

# Install swag + generate docs during build
# Generate Swagger (with error handling + fallback)
#RUN echo "=== Installing swag ===" && \
#    go install github.com/swaggo/swag/cmd/swag@latest && \
#    export PATH="$PATH:$(go env GOPATH)/bin" && \
#    echo "=== Swag version ===" && \
#    swag version && \
#    echo "=== Running swag init ===" && \
#    swag init --parseDependency --parseInternal --parseDepth 10 --generalApiInfo true && \
#    echo "=== SUCCESS: Docs generated ===" && \
#    ls -la docs/ && head -c 500 docs/swagger.json

RUN go test ./... -v -coverprofile=coverage.out -timeout 2m

# Build directly - Go will auto-download missing deps during build
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./api

FROM alpine:3.20
WORKDIR /app
COPY --from=build /bin/api /bin/api

ENV PORT=8080
EXPOSE 8080

CMD ["/bin/api"]
