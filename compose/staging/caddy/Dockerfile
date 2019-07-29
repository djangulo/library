FROM node:12.6-alpine as builder

WORKDIR /client

# add `/app/node_modules/.bin` to $PATH
ENV PATH /client/node_modules/.bin:$PATH

# install and cache app dependencies
COPY ./client .
RUN yarn install
RUN yarn build

FROM alpine:latest


FROM abiosoft/caddy:latest
# FROM alpine:latest

RUN addgroup -S client \
    && adduser -S -G client client

RUN apk update \
    && apk --no-cache add ca-certificates

COPY ./compose/staging/caddy/Caddyfile /etc/Caddyfile

COPY --from=builder /client/build /build
RUN chown -R client:client /build