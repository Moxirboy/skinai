# ---- build ----
FROM golang:1.23-alpine AS build
WORKDIR /src

# deps first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# copy the rest (main.go is in root)
COPY . .

# build a single binary (main package is at module root)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server .

# ---- run ----
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
ENV PORT=8080
CMD ["/app/server"]
