FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o nexus ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/nexus /nexus
EXPOSE 4317
ENTRYPOINT ["/nexus"]
