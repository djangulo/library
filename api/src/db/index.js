const pgp = require('pg-promise')({
  capSQL: true
});
const initdb = require('./initdb');
const seeddb = require('./seeddb');
const config = require('../config');

const connStr = (dbName, dbHost, dbPass, dbUser, dbPort) => {
  return (
    'postgres://' +
    dbUser +
    ':' +
    dbPass +
    '@' +
    dbHost +
    ':' +
    dbPort +
    '/' +
    dbName
  );
};

/** Creates a database for the server
 * @typedef Datadirs
 * @type {Object}
 * @property {string} migrations Migrations directory. Required if migrate=true
 * @property {string} seed Seed data directory. Required if seed=true
 * @property {string} corpora Corpora directory. Required if seed=true
 *
 * @typedef CreateDbOptions
 * @type {Object}
 * @property {boolean} migrate Initialize DB, see ./initdb.js
 * @property {boolean} seed Seed DB, see ./seeddb.js
 * @property {string[]} pre Array of SQL statements to run before migration
 * @property {Datadirs} datadirs Data directories object
 *
 * @param {string} name Database name
 * @param {string} host Database host
 * @param {string} pass Database password
 * @param {string} user Database user
 * @param {string} port Database port
 * @param {CreateDbOptions} [options={}] options object
 * @returns {pgp.Database} db pg-promise Database
 */
const createDb = async (name, host, pass, user, port, options = {}) => {
  const cn = connStr(name, host, pass, user, port);
  const db = await pgp(cn);
  const {
    migrate,
    seed,
    pre,
    datadirs: { migrations: dMig, seed: dSeed, corpora: dCorp }
  } = options;
  if (migrate && !dMig) {
    throw new Error('Cannot initialize without a migrations dir.');
  }
  await (async function() {
    if (seed && (!dSeed || !dCorp)) {
      throw new Error('Cannot seed without corpora and seed dirs.');
    }
    if (migrate && seed) {
      await initdb(db, dMig, pre)
        .then(() => seeddb(db, dCorp, dSeed))
        .catch(err => {
          console.log('ERROR initializing db: ', err);
          throw err;
        });
    } else if (migrate && !seed) {
      await initdb(db, dMig);
    } else if (migrate && !seed) {
      await seeddb(db, dCorp, dSeed);
    }
  })();
  return db;
};

const { dbName, dbUser, dbHost, dbPass, dbPort } = config;

let cn = connStr(dbName, dbHost, dbPass, dbUser, dbPort);
const db = pgp(cn);

module.exports = { createDb, connStr, db };
