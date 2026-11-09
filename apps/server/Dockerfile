# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS base
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct \
    CGO_ENABLED=0
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS dev
RUN GOFLAGS=-p=1 go install github.com/air-verse/air@latest
ENV PATH="/go/bin:${PATH}" \
    GIN_MODE=debug
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM base AS build
RUN go build -p 1 -tags nomsgpack -trimpath -ldflags="-s -w" -o /out/todo-server .

FROM alpine:3.20 AS prod
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/todo-server .
ENV GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["./todo-server"]
