const uuidv4 = require('uuid/v4');
const db = require('./index').db;

const list = limit => {
  return db.any('SELECT * FROM books LIMIT $1;', [limit]);
};

const getById = bookId => {
  return db.any('SELECT * FROM books WHERE id = $1', [bookId]);
};

const search = (q, orderBy, order) => {
  return db.any(
    `
    SELECT * FROM books 
    WHERE title = $1 
    OR author = $1
    ORDER BY $2 $3;
  `,
    [q, orderBy, order]
  );
};

const create = (db, book) => {
  return db
    .one(
      `INSERT INTO books(
      id,
      title,
      slug,
      author,
      synopsis,
      pub_year,
      page_count
    ) VALUES($1, $2, $3, $4, $5, $6, $7);
    `,
      [
        uuidv4(),
        book.title,
        book.slug ? book.slug : null,
        book.author ? book.author : null,
        book.synopsis ? book.synopsis : null,
        book.pub_year ? book.pub_year : null,
        book.page_count ? book.page_count : null
      ]
    )
    .then(result => result.id)
    .error(error => console.log('ERROR: ', error));
};

module.exports = {
  list,
  search,
  getById,
  create
};
