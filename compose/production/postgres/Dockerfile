FROM postgres:11.3

COPY ./compose/production/postgres/maintenance /usr/local/bin/maintenance
RUN chmod +x /usr/local/bin/maintenance/*
RUN mv /usr/local/bin/maintenance/* /usr/local/bin \
    && rmdir /usr/local/bin/maintenance

RUN mkdir -p /docker-entrypoint-initdb.d

COPY ./compose/production/postgres/init.sh /docker-entrypoint-initdb.d/init.sh
