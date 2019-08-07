const cors = require('cors');

const express = require('express');

const createServer = require('./index');
const createDb = require('./db/index').createDb;
const config = require('./config');
const routes = require('./routes');

const bookDb = require('./db/book');
const pageDb = require('./db/page');

const {
  dbName,
  dbUser,
  dbHost,
  dbPass,
  dbPort,
  datadirs,
  clientAddr,
  servePort,
  rootUrl
} = config;

const dbOptions = {
  migrate: true,
  seed: true,
  datadirs
};

const appRoutes = [
  { path: '/books', router: routes.books },
  { path: '/pages', router: routes.pages }
];

var db;
createDb(dbName, dbHost, dbPass, dbUser, dbPort, dbOptions).then(
  dtbs => (db = dtbs)
);
const context = {
  db,
  bookDb,
  pageDb
};
const app = express();
app.use(cors());
app.use((req, res, next) => {
  next();
});
for (let route of appRoutes) {
  app.use(route.path, route.router);
}
// createServer([cors], appRoutes, context).then(app => {
//
app.get('/', (req, res) => {
  res.redirect(301, clientAddr);
});

app.listen(servePort, () =>
  console.log(
    'Library API listening on ' +
      (rootUrl ? rootUrl + ':' + servePort : 'port :' + servePort)
  )
);
// });
