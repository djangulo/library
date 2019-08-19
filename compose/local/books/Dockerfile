FROM golang:1.12.8 as builder
LABEL maintainer="Denis Angulo <djal@tuta.io>"
ENV GO111MODULE=on

# tool to wait for postgres connection
RUN git clone https://github.com/eficode/wait-for.git /wait-for

# install the `migrate` tool to manage migrations
# https://github.com/golang-migrate/migrate
RUN git clone https://github.com/golang-migrate/migrate.git /migrate
WORKDIR /migrate
RUN git checkout v4.5.0 \
    && CGO_ENABLED=0 go build \
    -tags 'postgres' \
    -ldflags="-X main.Version=$(git describe --tags)" \
    -o $GOPATH/bin/migrate \
    /migrate/cmd/migrate


WORKDIR $GOPATH/src/github.com/djangulo/library


# copy go module files to cache dependencies
COPY ./go.mod .
COPY ./go.sum .

RUN go get -d -u

# Copy only necesary files
COPY ./books/ ./books/
COPY ./config/ ./config/
COPY ./cmd/ ./cmd/


# create compiled test binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go test ./books \
    -c \
    -a -v \
    -installsuffix cgo \
    -o $GOPATH/bin/books.test

# compile app binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -tags netgo \
    -p 1 \
    -a \
    -installsuffix cgo \
    -o $GOPATH/bin/books \
    ./cmd/books/main.go

# Copy compose, data, migrations and static after compile to prevent
# unnecesary compilations by non-essential changes
COPY ./static/ ./static/
COPY ./data/ ./data/
COPY ./migrations/ ./migrations/
COPY ./compose/ ./compose/



# FROM scratch
FROM alpine:latest


RUN addgroup -S golang \
    && adduser -S -G golang golang

RUN apk update \
    && apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /go/bin/books /books
COPY --from=builder /go/bin/books.test /test
COPY --from=builder /go/bin/migrate /migrate
COPY --from=builder /wait-for/wait-for /wait-for
COPY --from=builder /go/src/github.com/djangulo/library/compose/local/books/entrypoint /entrypoint
COPY --from=builder /go/src/github.com/djangulo/library/compose/local/books/start /start
COPY --from=builder /go/src/github.com/djangulo/library/data/ /appdata/
COPY --from=builder /go/src/github.com/djangulo/library/migrations/ /migrations/
COPY --from=builder /go/src/github.com/djangulo/library/static/ /static/
# RUN sed -i 's/\r//g' /wait-for
RUN chmod +x /test /books /wait-for /start /entrypoint /migrate
RUN chown -R golang:golang \
    /books \
    /test \
    /migrate \
    /wait-for \
    /entrypoint \
    /start \
    /appdata \
    /migrations \
    /static

USER golang
WORKDIR /

ENTRYPOINT ["/entrypoint"]