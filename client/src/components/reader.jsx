/** @jsx jsx */
// eslint-disable-next-line
import React from 'react';
import { jsx } from '@emotion/core';
import ReaderHeader from './readerHeader';
import ReaderBody from './readerBody';

import { connect } from 'react-redux';

import { Helmet } from 'react-helmet';
import {
  setPage,
  lastPageByBook,
  getBooks,
  selectedBook,
  selectBook
} from '../store';

const Reader = ({
  book,
  books,
  lastPages,
  setPage,
  match,
  location,
  selectBook
}) => {
  // console.log(location);
  // console.log(match);

  // React.useEffect(() => {
  //   const { slug } = match.params;
  //   let book;
  //   if (slug) {
  //     book = books.find(b => b.slug === slug);
  //     selectBook(book);
  //   } else {
  //     selectBook(null);
  //   }
  // }, [books, match.params, selectBook]);

  // React.useEffect(() => {
  //   const qParams = new URLSearchParams(location.search);
  //   const page = qParams.get('page');

  //   if (page) {
  //     setPage(book.id, page);
  //   } else {
  //     setPage(lastPages[book.id] ? lastPages[book.id] : 1);
  //   }
  // }, [book.id, lastPages, location.search, setPage]);

  const title = book =>
    `Library${
      book && book.title ? ` | Read ${book.title} by ${book.author}` : ''
    }`;
  // console.log(match);
  return book ? (
    <div id="reader" css={{ margin: 'auto' }}>
      <Helmet>
        <title>{title()}</title>
      </Helmet>
      <ReaderHeader />
      <ReaderBody />
    </div>
  ) : null;
};

export default connect(
  state => ({
    lastPages: lastPageByBook(state),
    books: getBooks(state),
    book: selectedBook(state)
  }),
  { setPage, selectBook }
)(Reader);
