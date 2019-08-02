import uuidv4 from 'uuid/v4';
// import * as pages from './data/fakePages.json';

const pages = require('../data/fakePages.json');

export const list = () =>
  new Promise(resolve => setTimeout(() => resolve(pages), 400));

export const retrieve = (bookId, pageNumber) =>
  new Promise(resolve =>
    setTimeout(
      () =>
        resolve(pages.find(p => p.book === bookId && p.number === pageNumber)),
      400
    )
  );

export const update = page =>
  new Promise(resolve =>
    setTimeout(
      () =>
        resolve(() => {
          let dbPage = pages.find(p => p.id === page.id) || {};
          if (page.book) dbPage.book = page.book;
          if (page.number) dbPage.number = page.number;
          if (page.text) dbPage.text = page.text;

          if (!dbPage.id) {
            dbPage.id = uuidv4().toString();
            pages.push(dbPage);
          }

          return dbPage;
        }),
      400
    )
  );

export const remove = id =>
  new Promise(resolve =>
    setTimeout(
      () =>
        resolve(() => {
          let dbPage = pages.find(p => p.id === id);
          pages.splice(pages.indexOf(dbPage), 1);
          return dbPage;
        }),
      400
    )
  );

export default {
  remove,
  list,
  retrieve,
  update
};
