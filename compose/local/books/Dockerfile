FROM golang:1.12.8 as builder
LABEL maintainer="Denis Angulo <djal@tuta.io>"
ENV GO111MODULE=on
WORKDIR $GOPATH/src/github.com/djangulo/library/

# copy dependency management to cache dependencies
COPY ./go.mod .
COPY ./go.sum .

RUN go get -d -u

RUN mkdir ./static
COPY ./static/books ./static/books
COPY ./books ./books
COPY ./config ./config
COPY ./migrations ./migrations
COPY ./data ./data
RUN mkdir ./cmd
COPY ./cmd/books ./cmd/books

# create compiled  test binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go test ./books \
    -c \
    -a \
    -installsuffix cgo \
    -o $GOPATH/bin/books.test

# compile app binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -p 1 \
    -a \
    -installsuffix cgo \
    -o $GOPATH/bin/books \
    ./cmd/books/main.go


# tool to wait for postgres connection
RUN git clone https://github.com/eficode/wait-for.git /wait-for



# FROM scratch
FROM alpine:latest

RUN addgroup -S golang \
    && adduser -S -G golang golang

RUN apk update \
    && apk --no-cache add ca-certificates

COPY --from=builder /go/bin/books /books
RUN chmod +x ./books
RUN chown golang:golang /books

COPY --from=builder /go/bin/books.test /test
RUN chmod +x ./test
RUN chown golang:golang /test

COPY --from=builder /go/src/github.com/djangulo/library/static/books/html /static/html
COPY --from=builder /go/src/github.com/djangulo/library/migrations /migrations
COPY --from=builder /go/src/github.com/djangulo/library/data /data
RUN chown -R golang:golang /static /data /migrations

COPY --from=builder /wait-for/wait-for /wait-for
RUN sed -i 's/\r//g' /wait-for
RUN chmod +x /wait-for
RUN chown golang:golang /wait-for

USER golang