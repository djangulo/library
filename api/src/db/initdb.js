const fsPromises = require('fs').promises;
const fs = require('fs');

const seed = require('./seeddb');

/**
 * Runs migrations after initializing the database
 * @param {*} db pg-promise db instance
 * @param {string} migrationsDir dir where migrations live
 * @param {string[]} pre array of SQL statements to run prior to the migrations
 */
const runMigrations = async (db, migrationsDir, pre) =>
  await fsPromises
    .readdir(migrationsDir, { encoding: 'utf8' })
    .then(async migrations => {
      if (pre.length) {
        for (let stmt of pre) {
          console.log('Executing statement: ' + stmt + '...');
          await db
            .any(stmt)
            .then(console.log('SUCCESS: ' + stmt))
            .catch(e => {
              console.log('\nERROR executing statement ' + stmt + ':\n');
              throw e;
            });
        }
      }
      for (let migration of migrations) {
        if (migration.indexOf('README') !== -1) continue;
        const data = fs.readFileSync(migrationsDir + '/' + migration, 'utf8');
        console.log('Executing migration: ' + migration + '...');
        await db
          .any(data)
          .then(() => console.log('SUCCESS: ' + migration))
          .catch(e => {
            console.log('\nERROR executing migration ' + migration + ':\n');
            throw e;
          });
      }
    })
    .catch(e => {
      console.log('\nERROR reading dir ' + migrationsDir);
      throw e;
    });

const initdb = async (db, migrationsDir, pre = []) =>
  await runMigrations(db, migrationsDir, pre);

module.exports = initdb;
