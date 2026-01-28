# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o skills ./cmd/skills

# Final stage
FROM scratch

COPY --from=builder /build/skills /usr/local/bin/skills

ENTRYPOINT ["/usr/local/bin/skills"]
