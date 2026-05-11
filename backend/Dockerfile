ARG GO_VERSION=1.26.1

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

COPY go.work ./
COPY backend/auth/go.mod backend/auth/go.sum backend/auth/
COPY backend/profile/go.mod backend/profile/go.sum backend/profile/
COPY backend/user/go.mod backend/user/go.sum backend/user/
COPY backend/gateway/go.mod backend/gateway/
COPY test/bdd/go.mod test/bdd/go.sum test/bdd/
COPY test/smoke/go.mod test/smoke/

RUN go work sync

COPY . .

ARG SERVICE
RUN CGO_ENABLED=0 go build -o /out/server ./backend/${SERVICE}/cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=build /out/server /usr/local/bin/server

ENTRYPOINT [ "/usr/local/bin/server" ]
