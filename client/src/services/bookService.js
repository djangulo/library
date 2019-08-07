// import uuidv4 from 'uuid/v4';
import config from '../config';

const { apiRoot } = config;

export const list = () => fetch(`${apiRoot}/books`);

export const retrieve = id => fetch(`${apiRoot}/books/${id}`);

// export const update = book =>
//   new Promise(resolve =>
//     setTimeout(
//       () =>
//         resolve(() => {
//           let bookInDb = books.find(b => b.id === book.id) || {};
//           if (book.title) bookInDb.title = book.title;
//           if (book.slug) bookInDb.slug = book.slug;
//           if (book.author) bookInDb.author = book.author;
//           if (book.synopsis) bookInDb.synopsis = book.synopsis;
//           if (book.date_added) bookInDb.date_added = book.date_added;
//           if (book.publication_date)
//             bookInDb.publication_date = book.publication_date;
//           if (book.isbn) bookInDb.isbn = book.isbn;

//           if (!bookInDb.id) {
//             bookInDb.id = uuidv4().toString();
//             books.push(bookInDb);
//           }

//           return bookInDb;
//         }),
//       400
//     )
//   );

// export const remove = id =>
//   new Promise(resolve =>
//     setTimeout(
//       () =>
//         resolve(() => {
//           let bookInDb = books.find(b => b.id === id);
//           books.splice(books.indexOf(bookInDb), 1);
//           return bookInDb;
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
