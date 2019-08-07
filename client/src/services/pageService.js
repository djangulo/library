// import uuidv4 from 'uuid/v4';
import config from '../config';

const { apiRoot } = config;

export const list = () => fetch(`${apiRoot}/pages`);

export const retrieve = (bookId, pageNumber) =>
  fetch(
    `${apiRoot}/pages/by-params?book-id=${bookId}&page-number=${pageNumber}`
  );

// export const update = page =>
//   new Promise(resolve =>
//     setTimeout(
//       () =>
//         resolve(() => {
//           let dbPage = pages.find(p => p.id === page.id) || {};
//           if (page.book) dbPage.book = page.book;
//           if (page.number) dbPage.number = page.number;
//           if (page.text) dbPage.text = page.text;

//           if (!dbPage.id) {
//             dbPage.id = uuidv4().toString();
//             pages.push(dbPage);
//           }

//           return dbPage;
//         }),
//       400
//     )
//   );

// export const remove = id =>
//   new Promise(resolve =>
//     setTimeout(
//       () =>
//         resolve(() => {
//           let dbPage = pages.find(p => p.id === id);
//           pages.splice(pages.indexOf(dbPage), 1);
//           return dbPage;
//         }),
//       400
//     )
//   );

export default {
  // remove,
  list,
  retrieve
  // update
};
