#!/usr/bin/env bash


set -o errexit
set -o pipefail
# set -o nounset

if  [ -z POSTGRES_ROOTDIR ]; then
    POSTGRES_ROOTDIR=/var/lib/postgresql
fi

if  [ -z POSTGRES_USER ]; then
    POSTGRES_USER=postgres
fi

if  [ -z POSTGRES_PORT ]; then
    POSTGRES_PORT=5432
fi

if  [ -z POSTGRES_DB ]; then
    POSTGRES_DB=postgres
fi

openssl req \
    -new -text \
    -passout pass:$POSTGRES_SSLPASSPHRASE
    # -subj /CN=localhost # TODO needs research
    -out $POSTGRES_ROOTDIR/server.req

openssl rsa \
    -in privkey.pem \
    -passin pass:$POSTGRES_SSLPASSPHRASE \
    -out $POSTGRES_ROOTDIR/server.key

openssl req -x509 \
    -in $POSTGRES_ROOTDIR/server.req \
    -text -key $POSTGRES_ROOTDIR/server.key \
    -out $POSTGRES_ROOTDIR/server.crt

openssl dhparam \
    -out $POSTGRES_ROOTDIR/dhparams.pem 4096

chown $POSTGRES_USER:$POSTGRES_USER \
    $POSTGRES_ROOTDIR/server.key

chmod 600 $POSTGRES_ROOTDIR/server.key

cat > $POSTGRES_ROOTDIR/secure_pg_hba.conf << EOF
# allow only access to the postgres user from within the docker network
hostssl $POSTGRES_DB $POSTGRES_USER 172.18.0.0/16 scram-sha-256

EOF

cat > $POSTGRES_ROOTDIR/secure.conf << EOF
hba_file = 'secure_pg_hba.conf'

# TODO: inject unique container addresses (see https://stackoverflow.com/a/20686101)
listen_addresses = '*'

# obscured port
port = $POSTGRES_PORT

# swap md5 default for stronger sha-256 option
password_encryption = scram-sha-256

# ssl config
ssl = on
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'
ssl_dh_params_file = 'dhparams.pem'
ssl_passphrase_command = 'echo $POSTGRES_SSLPASSPHRASE'

EOF

# append new configs to main config file
echo "include 'secure.conf'" >> $POSTGRES_ROOTDIR/postgresql.conf