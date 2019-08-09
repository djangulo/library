const Router = require('express').Router;

const pageDb = require('../db/page');
const isPaginated = require('../config').paginate;
const paginatedResponse = require('../utils/paginateResponse');

const router = Router();

router.get('/by-params', (req, res) => {
  const bookId = req.query['book-id'];
  const pageNumber = req.query['page-number'];
  if (!pageNumber || !bookId) return [];
  return pageDb
    .getByParams(bookId, pageNumber)
    .then(data => {
      return res.send(data);
    })
    .catch(err => console.log(err));
});

router.get('/:id', (req, res) => {
  return pageDb
    .get(req.params.id)
    .then(data => {
      return res.send(data);
    })
    .catch(err => console.log(err));
});

router.get('/', (req, res) => {
  if (isPaginated) {
    let { page = 1 } = req.query;
    return pageDb
      .list(-1)
      .then(data => {
        const paginated = paginatedResponse(
          data.length,
          '/pages',
          parseInt(page, 10),
          data
        );
        return res.send(paginated);
      })
      .catch(err => console.log(err));
  }
  return pageDb
    .list(1000)
    .then(data => res.send(data))
    .catch(err => console.log(err));
});

module.exports = router;
