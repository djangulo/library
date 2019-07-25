FROM node:12.6-alpine

WORKDIR /client

# add `/app/node_modules/.bin` to $PATH
ENV PATH /client/node_modules/.bin:$PATH

# install and cache app dependencies
COPY ./client .
RUN yarn install
