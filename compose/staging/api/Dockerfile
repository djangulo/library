FROM alpine:latest

WORKDIR /api

COPY api/package.json .
COPY api/yarn.lock .


RUN apk update && apk add --virtual --no-cache git python3 nodejs yarn\
    && git clone https://github.com/eficode/wait-for.git wait \
    && python3 -m pip install --upgrade pip \
    && python3 -m pip install nltk \
    && mkdir -p ./src/db/nltk_data \
    && python3 -m nltk.downloader -d ./src/db/nltk_data gutenberg \
    && python3 -m pip uninstall nltk \
    && apk del python3 git

RUN yarn install

RUN addgroup -S api \
    && adduser -S -G api api

RUN cp ./wait/wait-for .
RUN sed -i 's/\r//g' ./wait-for
RUN chmod +x ./wait-for
RUN chown api:api ./wait-for


COPY ./api .


RUN chown -R api:api .



USER api
