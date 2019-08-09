/** @jsx jsx */
import React from 'react';
import { jsx } from '@emotion/core';
import styled from '@emotion/styled';

import { connect } from 'react-redux';

import { Route, withRouter } from 'react-router-dom';

import { Helmet } from 'react-helmet';

import Books from './components/books';

import { Sidebar, Header, Segment } from 'semantic-ui-react';
import {
  selectedBook,
  selectedPage,
  setPage,
  fetchBooks,
  lastPageByBook,
  getBooks,
  selectBook
} from './store';
import Reader from './components/reader';

const StyledSegment = styled(Segment)`
  height: 100vh;
`;

const HeaderComponent = styled(Segment)`
  height: 8em;
  margin: 0;
  display: flex;
  align-items: center;
  background: #006060;
`;

const App = ({
  book,
  books,
  selectedPage,
  match,
  location,
  setPage,
  fetchBooks,
  lastPages,
  selectBook
}) => {
  const [visible] = React.useState(true);

  React.useEffect(() => {
    const asyncFetch = async () => {
      fetchBooks(1);
      // const { slug } = match.params;
      // if (slug && books) {
      //   const book = books.find(b => b.slug === slug);
      //   if (book) selectBook(book);
      // }
    };
    asyncFetch();
  }, [books, fetchBooks, match.params, selectBook]);

  // React.useEffect(() => {
  //   const qParams = new URLSearchParams(location.search);
  //   const page = qParams.get('page');

  //   if (book && book.id && page) {
  //     setPage(book.id, page);
  //   } else if (book && book.id) {
  //     setPage(
  //       book.id,
  //       lastPages && lastPages[book.id] ? lastPages[book.id] : 1
  //     );
  //   }
  // }, [book, book.id, lastPages, location.search, match.params, setPage]);
  return (
    <div className="App">
      <Helmet>
        <title>Library</title>
      </Helmet>
      <HeaderComponent className="App-header">
        <Header as="h2" floated="left">
          <Header.Content as="a" href="/">
            Library
          </Header.Content>
        </Header>
      </HeaderComponent>
      <Sidebar.Pushable as={StyledSegment} css={{ marginTop: 0 }}>
        <Sidebar
          as={Segment}
          animation="slide out"
          icon="labeled"
          vertical
          visible={visible}
          width="wide"
        >
          <Books />
        </Sidebar>

        <Sidebar.Pusher
          as={StyledSegment}
          css={{
            maxWidth: '820px',
            paddingTop: 0,
            margin: '0 !important'
          }}
          basic
        >
          {book ? (
            <Route
              path="/books/:slug"
              component={props => <Reader {...props} />}
            />
          ) : null}
        </Sidebar.Pusher>
      </Sidebar.Pushable>
    </div>
  );
};

export default withRouter(
  connect(
    state => ({
      book: selectedBook(state),
      selectedPage: selectedPage(state),
      lastPages: lastPageByBook(state),
      books: getBooks(state)
    }),
    { setPage, fetchBooks, selectBook }
  )(App)
);
