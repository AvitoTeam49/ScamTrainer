FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/app ./cmd/app

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app

COPY --from=build /out/app /usr/local/bin/app
COPY scenarios /scenarios

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/app"]
