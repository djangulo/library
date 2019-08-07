export const books = require('./fakeData/fakeBooks.json');
export const pages = require('./fakeData/fakePages.json');

const createStore = (dbName, dbHost, dbPass, dbUser, dbPort, options) => {
  const connStr = `postgres://${dbUser}:${dbPass}@${dbHost}:${dbPort}/${dbName}`;
};
