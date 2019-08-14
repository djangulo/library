FROM golang:1.12.8 as builder
LABEL maintainer="Denis Angulo <djal@tuta.io>"
WORKDIR $GOPATH/src/github.com/djangulo/library/
COPY . .
RUN go get -d -v ./...

# compile app binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -p 1 \
    -a \
    -installsuffix cgo \
    -o $GOPATH/bin/books \
    ./cmd/books/main.go

# create compiled  test binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go test \
    -c \
    -a \
    -installsuffix cgo \
    -o $GOPATH/bin/books.test

# tool to wait for postgres connection
RUN git clone https://github.com/eficode/wait-for.git /wait-for



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

COPY --from=builder /go/src/github.com/djangulo/library/static/books /static
RUN chown -R golang:golang /html

COPY --from=builder /wait-for/wait-for /wait-for
RUN sed -i 's/\r//g' /wait-for
RUN chmod +x /wait-for
RUN chown golang:golang /wait-for
