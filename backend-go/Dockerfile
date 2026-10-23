# syntax=docker/dockerfile:1.7

# Etapa de build: compila el binario estático.
FROM golang:1.27.3-bookworm AS build
WORKDIR /src

# Capa de dependencias (mejora de cache).
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Etapa de runtime: imagen mínima, sin shell, sin privilegios.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
