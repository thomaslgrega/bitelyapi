# Stage 1:
FROM golang:1.24.0-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2:
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/server /server

ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "/server" ]