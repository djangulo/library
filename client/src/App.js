/** @jsx jsx */
import React from 'react';
import { jsx } from '@emotion/core';
import styled from '@emotion/styled';

import { connect } from 'react-redux';

import { Route, withRouter } from 'react-router-dom';

import { Helmet } from 'react-helmet';

import Books from './components/books';

import { Sidebar, Header, Segment } from 'semantic-ui-react';
import { selectedBook, fetchBooks } from './store';
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

const App = ({ book, fetchBooks }) => {
  const [visible] = React.useState(true);

  React.useEffect(() => {
    fetchBooks(1);
  }, [fetchBooks]);

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
            maxWidth: '920px',
            paddingTop: 0,
            margin: '0 !important',
            overflow: 'scroll !important'
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
      book: selectedBook(state)
    }),
    { fetchBooks }
  )(App)
);
