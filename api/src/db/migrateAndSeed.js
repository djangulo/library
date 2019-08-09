var pgsync = require('pg-sync');
var client = new pgsync.Client();
const initdb = require('./initdb');
const seeddb = require('./seeddb').seeddb;
const config = require('../config');

const { dbUser, dbPass, dbHost, dbPort, dbName, datadirs } = config;

const cn =
  'postgres://' +
  dbUser +
  ':' +
  dbPass +
  '@' +
  dbHost +
  ':' +
  dbPort +
  '/' +
  dbName;

const db = pgp(cn);

initdb(db, datadirs.migrations);
seeddb(db, datadirs.seed);
