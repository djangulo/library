const cors = require('cors');

const express = require('express');
const path = require('path');

const config = require('./config');
const routes = require('./routes');

const { servePort, rootUrl } = config;

const app = express();
app.use(cors());
app.use('/books', routes.books);
app.use('/pages', routes.pages);

app.get('/', (req, res) => {
  res.redirect(301, '/en');
});

app.get('/en', (req, res) => {
  res.sendFile(path.join(__dirname + '/index.en.html'));
});

app.get('/es', (req, res) => {
  res.sendFile(path.join(__dirname + '/index.es.html'));
});

app.listen(servePort, () =>
  console.log(
    'Library API listening on ' + (rootUrl ? rootUrl : 'port :' + servePort)
  )
);
