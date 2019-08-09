const _ = require('lodash');
const Router = require('express').Router;

const bookDb = require('../db/book');

const isPaginated = require('../config').paginate;
const paginatedResponse = require('../utils/paginateResponse');

const router = Router();

const validateOrder = order => {
  if (order === 'asc') return 'asc';
  if (order === 'desc') return 'desc';
  return 'asc';
};
const columns = [
  'title',
  'slug',
  'author',
  'pub_year',
  'id',
  'page_count',
  'file'
];
const validateOrderBy = orderBy => {
  for (let column of columns) {
    if (orderBy === column) return orderBy;
  }
  return 'title';
};

router.get('/search', (req, res) => {
  if (!req.query.q) return res.send([]);
  const { q, sort = 'title', order = 'asc', page = 1 } = req.query;
  return bookDb
    .list(-1)
    .then(data => {
      const filtered = data.filter(
        b =>
          b.title.toLowerCase().includes(q.toLowerCase()) ||
          (b.author && b.author.toLowerCase().includes(q.toLowerCase()))
      );
      const ordered = _.orderBy(
        filtered,
        [validateOrderBy(sort)],
        [validateOrder(order)]
      );
      const paginated = paginatedResponse(
        ordered.length,
        '/books/search',
        page,
        ordered
      );
      return res.send(paginated);
    })
    .catch(err => console.log(err));
});

router.get('/', (req, res) => {
  if (isPaginated) {
    let { page = 1 } = req.query;
    return bookDb
      .list(-1)
      .then(data => {
        const paginated = paginatedResponse(
          data.length,
          '/books',
          parseInt(page, 10),
          data
        );
        return res.send(paginated);
      })
      .catch(err => console.log(err));
  }
  return bookDb
    .list(1000)
    .then(data => res.send(data))
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
