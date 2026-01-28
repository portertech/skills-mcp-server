# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o skills ./cmd/skills

# Final stage
FROM scratch

COPY --from=builder /build/skills /usr/local/bin/skills

ENTRYPOINT ["/usr/local/bin/skills"]
