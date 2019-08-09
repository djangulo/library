import React from 'react';

import { connect } from 'react-redux';

import { Header, Message } from 'semantic-ui-react';
import { selectedBook, selectedPage, pagesError } from './../store';
import ReaderMenu from './readerMenu';

const ReaderHeader = ({ book, pagesErr }) => {
  return book ? (
    <div>
      <Header as="h2">
        <Header.Content>
          {book.title}
          <Header.Subheader>{book.author}</Header.Subheader>
        </Header.Content>
      </Header>
      {pagesErr ? (
        <Message negative className="error">
          <p>{pagesErr}</p>
        </Message>
      ) : null}
      <ReaderMenu />
    </div>
  ) : null;
};

export default connect(state => ({
  book: selectedBook(state),
  page: selectedPage(state),
  pagesErr: pagesError(state)
}))(ReaderHeader);
