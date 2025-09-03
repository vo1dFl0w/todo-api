FROM golang:1.24.3

WORKDIR /todo-api

# ---- dependency caching ----
COPY go.mod go.sum ./
RUN go mod download

# ---- copying source files
COPY ./ ./

# ---- building a static optimized binary file ----
ENV CONFIG_PATH=config/config.yml
RUN go build -ldflags="-s -w" -o todo-app ./cmd/todo-api/main.go

EXPOSE 8080

CMD ["./todo-app"]

