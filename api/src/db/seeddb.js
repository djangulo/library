const fs = require('fs');
const fsPromises = require('fs').promises;
const readline = require('readline');
const pgp = require('pg-promise')({ capSQL: true });
// const execSync = require('child_process').execSync;
const uuidv4 = require('uuid/v4');

const config = require('../config');

// const db = require("./index");
const paginate = require('../utils/paginate').paginate;

// const corporaDir = "src/db/nltk_data/corpora/gutenberg";

// execSync('python3 ././init_nltk.py');

const slugify = str =>
  str
    .toLowerCase()
    .replace(/[._ ]/g, '-')
    .replace(/['\,]/, '');

const getMetadata = data => {
  const re = /\[([',-\w\s]+)\,? [bB]y ([.a-zA-Z\s]+)\ ?([\d]+)?\]/;
  const meta = data.replace('\r', '').match(re);
  if (meta) {
    const title = meta[1].trim().replace(',', '');
    return {
      title,
      slug: slugify(title),
      author: meta[2] ? meta[2].trim() : null,
      pub_year: meta[3] ? parseInt(meta[3], 10) : null
    };
  }
  return {
    title: data.replace(/[\]\[]+/g, ''),
    slug: slugify(data.replace(/[\]\[]+/g, '')),
    pub_year: null,
    author: null
  };
};

const readFirstLine = filePath => {
  return new Promise((resolve, reject) => {
    var rs = fs.createReadStream(filePath, { encoding: 'utf8' });
    var acc = '';
    var pos = 0;
    var index;
    rs.on('data', function(chunk) {
      index = chunk.indexOf('\n');
      acc += chunk;
      index !== -1 ? rs.close() : (pos += chunk.length);
    })
      .on('close', function() {
        resolve(acc.slice(0, pos + index));
      })
      .on('error', function(err) {
        reject(err);
      });
  });
};

const getBooks = async (corporaDir, seedDir) => {
  const booksPath = seedDir + '/books.json';
  if (fs.existsSync(seedDir) && fs.existsSync(booksPath)) {
    return fsPromises
      .readFile(booksPath)
      .then(buf => JSON.parse(buf))
      .catch(e => {
        console.log('\nERROR reading ' + booksPath + ':\n');
        throw e;
      });
  } else {
    // file does not exist, generate
    return fsPromises
      .readdir(corporaDir, 'utf8')
      .then(async files => {
        // create dir if not exists
        if (!fs.existsSync(seedDir)) fs.mkdirSync(seedDir);
        const stream = fs.createWriteStream(booksPath, { encoding: 'utf8' });
        stream.write('[\n');
        for (let i = 0; i < files.length; i++) {
          const file = files[i];
          if (file.indexOf('README') !== -1) continue;
          const filePath = corporaDir + '/' + files[i];

          await readFirstLine(filePath)
            .then(line => getMetadata(line))
            .then(b => ({
              ...b,
              id: uuidv4(),
              page_count: null,
              file
            }))
            .then(book =>
              stream.write(
                JSON.stringify(book, null, 2) +
                  (i === files.length - 1 ? '\n' : ',\n')
              )
            )
            .catch(e => {
              console.log('\nERROR reading book ', filePath);
              throw e;
            });
        }
        return stream;
      })
      .then(stream => {
        stream.write(']\n');
        return stream;
      })
      .then(stream => stream.end(() => console.log(booksPath + ' written.')))
      .then(() =>
        fsPromises
          .readFile(booksPath, 'utf8')
          .then(buf => JSON.parse(buf))
          .catch(e => {
            console.log('\nERROR reading ' + booksPath + ':\n');
            throw e;
          })
      )
      .catch(e => {
        console.log('\nERROR generating ' + booksPath + ' data :\n');
        throw e;
      });
  }
};

const getPages = async (corporaDir, seedDir, books) => {
  const pagesPath = seedDir + '/pages.json';
  if (fs.existsSync(seedDir) && fs.existsSync(pagesPath)) {
    return fsPromises
      .readFile(pagesPath)
      .then(buf => ({
        pages: JSON.parse(buf),
        books
      }))
      .catch(e => {
        console.log('\nERROR reading ' + pagesPath + ':\n');
        throw e;
      });
  } else {
    // file does not exist, generate
    return fsPromises
      .readdir(corporaDir, 'utf8')
      .then(files => {
        // create dir if not exists
        if (!fs.existsSync(seedDir)) fs.mkdirSync(seedDir);
        const stream = fs.createWriteStream(pagesPath, { encoding: 'utf8' });
        stream.write('[\n');
        for (let i = 0; i < files.length; i++) {
          const file = files[i];
          const book = books.find(b => b.file === file);
          if (files[i].indexOf('README') !== -1) continue;
          const filePath = corporaDir + '/' + files[i];

          const data = fs.readFileSync(filePath, 'utf8');
          const paragraphs = data.split('\n');
          const paragraphsPerPage = config.paragraphsPerPage;
          const page_count = Math.floor(paragraphs.length / paragraphsPerPage);
          book.page_count = page_count;
          for (let j = 1; j <= page_count; j++) {
            const page = {
              id: uuidv4(),
              body: paginate(paragraphs, j, paragraphsPerPage)
                .map(p => p.replace(/[\n\r]+/g, ' '))
                .join('\n'),
              page_number: j,
              book_id: book.id
            };

            stream.write(
              JSON.stringify(page, null, 2) +
                (i === files.length - 1 && j === page_count ? '\n' : ',\n')
            );
          }
        }
        fs.writeFile(
          seedDir + '/books.json',
          JSON.stringify(books, null, 2),
          'utf8',
          err => {
            if (err) {
              console.log('\n ERROR overwriting ' + seedDir + '/books.json');
              throw e;
            }
          }
        );
        return { books, stream };
      })
      .then(obj => {
        obj.stream.write(']\n');
        return obj;
      })
      .then(obj => {
        obj.stream.end(() => console.log(pagesPath + ' written.'));
        return obj.books;
      })
      .then(books =>
        fsPromises
          .readFile(pagesPath, 'utf8')
          .then(buf => ({
            pages: JSON.parse(buf),
            books
          }))
          .catch(e => {
            console.log('\nERROR reading ' + pagesPath + ':\n');
            throw e;
          })
      )
      .catch(e => {
        console.log('\nERROR generating ' + pagesPath + ' data :\n');
        throw e;
      });
  }
};

const seeddb = (db, corporaDir, seedDir) => {
  return getBooks(corporaDir, seedDir).then(books =>
    getPages(corporaDir, seedDir, books)
      .then(async data => {
        const { books, pages } = await data;

        const bookCs = new pgp.helpers.ColumnSet(
          ['id', 'title', 'slug', 'author', 'pub_year', 'page_count'],
          { table: 'books' }
        );

        const pageCs = new pgp.helpers.ColumnSet(
          ['id', 'page_number', 'body', 'book_id'],
          { table: 'pages' }
        );
        const insertBooks =
          pgp.helpers.insert(books, bookCs) + ' ON CONFLICT (id) DO NOTHING';
        const insertPages =
          pgp.helpers.insert(pages, pageCs) + ' ON CONFLICT (id) DO NOTHING';

        await db.none(insertBooks).then(() => db.none(insertPages));
        console.log('Succesfully seeded Database');
      })
      .catch(e => console.log('ERROR seeding database: ', e))
  );
};

module.exports = seeddb;
