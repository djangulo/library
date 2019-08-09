const fs = require('fs');
const Client = require('pg-native');
const pgp = require('pg-promise')({
  capSQL: true
});
const client = new Client();

const config = require('../config');

const { dbUser, dbPass, dbHost, dbPort, dbName, datadirs } = config;

client.connectSync(
  'user=' +
    dbUser +
    ' password=' +
    dbPass +
    ' host=' +
    dbHost +
    ' port=' +
    dbPort +
    ' dbname=' +
    dbName
);

// client.begin();
// client.setIsolationLevelSerializable();

const readData = path => {
  try {
    const file = fs.readFileSync(path, 'utf8');
    const data = JSON.parse(file);
    return data;
  } catch (e) {
    console.log('\nERROR reading ' + path + ':\n');
    throw e;
  }
};

const getBooks = seedDir => {
  return readData(seedDir + '/books.json');
};

const getPages = seedDir => {
  return readData(seedDir + '/pages.json');
};

const seeddb = (client, seedDir) => {
  console.log('Seeding database...');
  try {
    const books = getBooks(seedDir);
    const pages = getPages(seedDir);
    const bookCs = new pgp.helpers.ColumnSet(
      ['id', 'title', 'slug', 'author', 'pub_year', 'page_count', 'file'],
      { table: 'books' }
    );

    const pageCs = new pgp.helpers.ColumnSet(
      ['id', 'page_number', 'body', 'book_id'],
      { table: 'pages' }
    );
    const insertBooks =
      pgp.helpers.insert(books, bookCs) + ' ON CONFLICT (id) DO NOTHING';
    const insertPages =
      pgp.helpers.insert(pages, pageCs) + ' ON CONFLICT (id) DO NOTHING';
    console.log('Seeding books data...');
    client.prepareSync('seed_books', insertBooks, 0);
    client.executeSync('seed_books');
    console.log('Seeding pages data...');
    client.prepareSync('seed_pages', insertPages, 0);
    client.executeSync('seed_pages');
    console.log('Database seeded!');
  } catch (e) {
    console.log('ERROR seeding database: ', e);
    throw e;
  }
};

/**
 * Runs migrations after initializing the database
 * @param {*} db pg-promise db instance
 * @param {string} migrationsDir dir where migrations live
 * @param {string[]} pre array of SQL statements to run prior to the migrations
 */
const runMigrations = (client, migrationsDir) => {
  try {
    const migrations = fs.readdirSync(migrationsDir, 'utf8');

    for (let migration of migrations) {
      if (migration.indexOf('README') !== -1) continue;
      const data = fs.readFileSync(migrationsDir + '/' + migration, 'utf8');
      console.log('Executing migration: ' + migration + '...');
      client.prepareSync(migration, data, 0);
      try {
        client.executeSync(migration);
        console.log('SUCCESS: ' + migration);
      } catch (e) {
        console.log('\nERROR executing migration ' + migration + ':\n');
        throw e;
      }
    }
  } catch (e) {
    console.log('\nERROR reading dir ' + migrationsDir);
    throw e;
  }
};

runMigrations(client, datadirs.migrations);
seeddb(client, datadirs.seed);

// client.commit();
// client.disconnect();
