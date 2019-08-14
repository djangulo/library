const fs = require('fs');
const uuidv4 = require('uuid/v4');

const corpora = require('../config').datadirs.corpora;
const seed = require('../config').datadirs.seed;
const paragraphsPerPage = require('../config').paragraphsPerPage;
const paginate = require('../utils/paginate');

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

const generate = (corporaDir, seedDir, callback) => {
  const booksPath = seedDir + '/books.json';
  const pagesPath = seedDir + '/pages.json';
  if (
    fs.existsSync(seedDir) &&
    fs.existsSync(booksPath) &&
    fs.existsSync(pagesPath)
  ) {
    console.log('Data exists, skipping...');
    return;
  }
  if (!fs.existsSync(seedDir)) fs.mkdirSync(seedDir);
  console.log('Generating data from ' + corporaDir);

  return fs.readdir(corporaDir, 'utf8', (err, files) => {
    const bStream = fs.createWriteStream(booksPath, { encoding: 'utf8' });
    const pStream = fs.createWriteStream(pagesPath, { encoding: 'utf8' });
    bStream.write('[\n');
    pStream.write('[\n');
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      if (file.indexOf('README') !== -1) continue;
      const filePath = corporaDir + '/' + files[i];
      const data = fs.readFileSync(filePath, 'utf8');
      const paragraphs = data.split('\n');

      const page_count = Math.floor(paragraphs.length / paragraphsPerPage);

      const book = {
        ...getMetadata(paragraphs[0]),
        id: uuidv4(),
        page_count,
        file
      };
      bStream.write(
        JSON.stringify(book, null, 2) + (i === files.length - 1 ? '\n' : ',\n')
      );

      for (let j = 1; j <= page_count; j++) {
        const page = {
          id: uuidv4(),
          body: paginate(paragraphs, j, paragraphsPerPage)
            .map(p => p.replace(/[\n\r]+/g, ' '))
            .join('\n'),
          page_number: j,
          book_id: book.id
        };
        pStream.write(
          JSON.stringify(page, null, 2) +
            (i === files.length - 1 && j === page_count ? '\n' : ',\n')
        );
      }
    }
    bStream.write(']\n');
    bStream.end();
    pStream.write(']\n');
    pStream.end();
    console.log('Data generated succesfully');
  });
};

generate(corpora, seed);
