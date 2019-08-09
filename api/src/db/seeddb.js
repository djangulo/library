const fs = require('fs');
const https = require('https');
const pgp = require('pg-promise')({ capSQL: true });
const paginate = require('../utils/paginate');

const readData = path => {
  try {
    const file = fs.readFileSync(path, 'utf8');
    const data = JSON.parse(file);
    return data;
  } catch (e) {
    console.log('\nERROR reading ' + path + ':\n');
    throw e;
  }
};

const readOrDownload = (filename, url, readDir) => {
  const path = readDir + '/' + filename;
  if (fs.existsSync(readDir) && fs.existsSync(path)) {
    return readData(path);
  } else {
    if (!fs.existsSync(readDir)) fs.mkdirSync(readDir);
    console.log(path + ' does not exist, downloading from ' + url);
    const fh = fs.createWriteStream(path, { flags: 'w', encoding: 'utf8' });
    https.get(url, res => {
      res.pipe(
        fh,
        { end: 'false' }
      );
      res.on('end', () => {
        fh.end();
        console.log('Downloaded ' + path);
        return readData(path);
      });
      res.on('error', e => {
        console.log('\nERROR downloading ' + path + ':\n');
        throw e;
      });
    });
  }
};

const getBooks = seedDir => {
  return readData(seedDir + '/books.json');
};

const getPages = seedDir => {
  return readData(seedDir + '/pages.json');
};

const seeddb = (db, seedDir) => {
  console.log('Seeding database...');
  try {
    const books = getBooks(seedDir);
    const pages = getPages(seedDir);
    const bookCs = new pgp.helpers.ColumnSet(
      ['id', 'title', 'slug', 'author', 'pub_year', 'page_count', 'file'],
      { table: 'books' }
    );

    const pageCs = new pgp.helpers.ColumnSet(
      ['id', 'page_number', 'body', 'book_id'],
      { table: 'pages' }
    );
    const booksTotal = Math.ceil(books.length / 100);
    const pagesTotal = Math.ceil(books.length / 100);
    for (let i = 1; i <= booksTotal; i++) {
      const sliced = paginate(books, i, 100);
      const insertBooks =
        pgp.helpers.insert(sliced, bookCs) + ' ON CONFLICT (id) DO NOTHING';
      db.none(insertBooks);
    }
    for (let i = 1; i <= pagesTotal; i++) {
      const sliced = paginate(pages, i, 100);
      const insertPages =
        pgp.helpers.insert(sliced, pageCs) + ' ON CONFLICT (id) DO NOTHING';
      db.none(insertPages);
    }
    console.log('Database seeded!');
  } catch (e) {
    console.log('ERROR seeding database: ', e);
    throw e;
  }
};

module.exports = { seeddb };
