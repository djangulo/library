const pgp = require("pg-promise")({
  capSQL: true
});
const config = require("../config");
const initdb = require("./initdb");
const seeddb = require("./seeddb");
const connStr = require("./db/index").connStr;
/**
 * Creates a testing database on the fly, to keep tests contained
 * @param {*} name
 * @param {*} host
 * @param {*} pass
 * @param {*} user
 * @param {*} port
 * @param {*} options
 */
const createTestDb = async () => {
  const mainCn = connStr(
    config.dbName,
    config.DbHost,
    config.DbPass,
    config.dbUser,
    config.DbPort
  );
  const mainDb = await pgp(mainCn);
  await mainDb
    .any("CREATE DATABASE library_test;")
    .then(async () => {})
    .catch(e => console.log("failed to create test database", e));
  const testCn = connStr(
    "library_test",
    config.DbHost,
    config.DbPass,
    config.dbUser,
    config.DbPort
  );

  const testDb = await pgp(testCn);
  await (async function() {
    await initdb(testDb, config.datadirs.migrations)
      .then(() => seeddb(testDb, config.datadirs.corpora, config.datadirs.seed))
      .catch(err => {
        console.log("ERROR initializing db: ", err);
        throw err;
      });
  })();
  const closeTestDb = () => {
    mainDb.any("DROP DATABASE library_test").finally(pgp.end);
  };
  return { testDb, closeTestDb };
};

module.exports = createTestDb;
