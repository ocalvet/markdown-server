FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY backend/go.mod backend/go.sum* ./
RUN go mod download || true

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o markdown-server .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/markdown-server .
COPY frontend/ ./frontend/

RUN mkdir -p ./markdown-files

ENV MARKDOWN_DIR=/app/markdown-files
ENV PORT=8703
ENV IGNORE_PATTERNS=""

EXPOSE 8703

CMD ["./markdown-server"]
