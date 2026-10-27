FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o basekit ./cmd/basekit

FROM scratch
COPY --from=builder /app/basekit /basekit
EXPOSE 7700
ENTRYPOINT ["/basekit"]
