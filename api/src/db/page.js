const db = require("./index");
const list = limit => {
  if (limit === -1) return db.any("SELECT * FROM pages;");
  return db.any("SELECT * FROM pages LIMIT $1;", [limit]);
};

const get = pageId => {
  return db.any("SELECT * FROM pages WHERE id = $1", [pageId]);
};

const getByParams = (bookId, pageNumber) => {
  return db.any(
    "SELECT * FROM pages WHERE book_id = $1 AND page_number = $2 LIMIT 1;",
    [bookId, pageNumber]
  );
};

module.exports = {
  list,
  get,
  getByParams
};
