FROM golang:1.21-alpine AS base

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Development stage with Air for hot reload
FROM base AS dev

RUN go install github.com/air-verse/air@v1.52.3

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

# Production stage
FROM base AS prod

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

FROM alpine:3.19 AS production
COPY --from=prod /server /server
EXPOSE 8080
CMD ["/server"]
