FROM golang:1.24.3-alpine AS builder

WORKDIR /todo-api

RUN apk --no-cache add bash git make gcc gettext musl-dev

# ---- dependency caching ----
COPY go.mod go.sum ./
RUN go mod download

# ---- copying source files
COPY ./ ./

# ---- building a static optimized binary file ----
ENV CONFIG_PATH=config/config.yml
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o todo-app ./cmd/todo-api/main.go

# ----- Runtime -----
FROM alpine AS runner
RUN apk add --no-cache ca-certificates

WORKDIR /todo-api
COPY --from=builder /todo-api/todo-app /todo-api/todo-app
COPY --from=builder /todo-api/config /todo-api/config

EXPOSE 8080

CMD ["./todo-app"]

