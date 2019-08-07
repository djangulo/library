FROM node:12.6-alpine as builder

WORKDIR /client

COPY client/package.json .
COPY client/yarn.lock .

RUN yarn install

# add `/app/node_modules/.bin` to $PATH
ENV PATH /client/node_modules/.bin:$PATH

# install and cache app dependencies

RUN addgroup -S client \
    && adduser -S -G client client


COPY ./client .

RUN chown -R client:client .

USER client


# FROM alpine:latest


# FROM abiosoft/caddy:latest
# # FROM alpine:latest

# RUN addgroup -S client \
#     && adduser -S -G client client

# RUN apk update \
#     && apk --no-cache add ca-certificates

# COPY ./compose/staging/caddy/Caddyfile /etc/Caddyfile

# COPY --from=builder /client/build /build
# RUN chown -R client:client /build