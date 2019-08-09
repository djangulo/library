const pgp = require('pg-promise')({
  capSQL: true
});
const config = require('../config');

const connStr = (user, pass, host, port, name) => {
  return (
    'postgres://' + user + ':' + pass + '@' + host + ':' + port + '/' + name
  );
};

const { dbUser, dbPass, dbHost, dbPort, dbName } = config;

const cn = connStr(dbUser, dbPass, dbHost, dbPort, dbName);

const db = pgp(cn);

module.exports = db;
