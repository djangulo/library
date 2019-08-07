const _ = require('lodash');
const Router = require('express').Router;

const bookDb = require('../db/book');
const db = require('../db/index').db;

const router = Router();

router.get('/search', (req, res) => {
  if (!req.query.q) return res.send([]);
  const { q, by = 'title', order = 'asc' } = req.query;

  return bookDb
    .search(q, by, order)
    .then(data => {
      return res.send(data);
    })
    .catch(err => console.log(err));
});

router.get('/', (req, res) => {
  console.log(req.context);
  return bookDb
    .list(1000)
    .then(data => {
      return res.send(data);
    })
    .catch(err => console.log(err));
});

router.get('/:id', (req, res) => {
  return bookDb
    .getById(req.params.id)
    .then(data => {
      return res.send(data);
    })
    .catch(err => console.log(err));
});

module.exports = router;
