const db = require('./index');

const list = limit => {
  if (limit === -1) return db.any('SELECT * FROM books;');
  return db.any('SELECT * FROM books LIMIT $1;', [limit]);
};

const getById = bookId => {
  return db.any('SELECT * FROM books WHERE id = $1', [bookId]);
};

module.exports = {
  list,
  getById
};
