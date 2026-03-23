FROM golang:1.26-alpine AS go-builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /adaptor ./cmd/adaptor

FROM python:3.13-slim AS final

RUN useradd -r -s /bin/false appuser

COPY --from=go-builder /adaptor /usr/local/bin/adaptor
COPY python/ /app/python/

WORKDIR /app
USER appuser

EXPOSE 8080 8090

COPY deployments/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
