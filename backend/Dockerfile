ARG GO_VERSION=1.26.1

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

COPY backend/. ./

ARG SERVICE
WORKDIR /src/${SERVICE}
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=build /out/server /usr/local/bin/server

ENTRYPOINT [ "/usr/local/bin/server" ]
