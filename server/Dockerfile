FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
ARG UNBIND_GIT_SHA
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=${UNBIND_GIT_SHA:-development}" -o /out/api ./cmd/api

FROM stablecog/ubuntu:22.04

COPY --from=builder /out/api /app/api
COPY ./blog /app/blog

ENV BLOG_DIR=/app/blog

EXPOSE 8337

CMD ["/app/api"]
