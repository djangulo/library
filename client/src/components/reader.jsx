/** @jsx jsx */
// eslint-disable-next-line
import React from 'react';
import { jsx } from '@emotion/core';
import ReaderHeader from './readerHeader';
import ReaderBody from './readerBody';

import { connect } from 'react-redux';

import { Helmet } from 'react-helmet';
import { selectedBook } from '../store';

const Reader = ({ book }) => {
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

export default connect(state => ({
  book: selectedBook(state)
}))(Reader);
