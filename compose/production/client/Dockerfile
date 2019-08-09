FROM node:12.6-alpine as builder

WORKDIR /client

COPY client/package.json .
COPY client/yarn.lock .

RUN yarn install && yarn global add serve

# add `/app/node_modules/.bin` to $PATH
ENV PATH /client/node_modules/.bin:$PATH

# install and cache app dependencies
COPY ./client .
RUN yarn build
RUN addgroup -S client \
    && adduser -S -G client client
RUN chown -R client:client /client

USER client
