/** @jsx jsx */
import React from 'react';
import { jsx } from '@emotion/core';
import styled from '@emotion/styled';

import { connect } from 'react-redux';

import BookList from './components/bookList';

import { Sidebar, Header, Segment } from 'semantic-ui-react';
import { selectedBook, selectedPage } from './store';
import Reader from './components/reader';

const StyledSegment = styled(Segment)`
  height: 100vh;
`;

const App = ({ selectedBook, selectedPage }) => {
  const [visible] = React.useState(true);
  return (
    <div className="App">
      <Header
        className="App-header"
        as="h1"
        css={{
          minHeight: '3em',
          display: 'flex',
          alignItems: 'center',
          marginBottom: 0
        }}
      >
        <Header.Content as="a" href="/">
          Library
        </Header.Content>
      </Header>
      <Sidebar.Pushable as={StyledSegment} css={{ marginTop: 0 }}>
        <Sidebar
          as={Segment}
          animation="slide out"
          icon="labeled"
          vertical
          visible={visible}
          width="wide"
        >
          <BookList />
        </Sidebar>

        {selectedBook && selectedBook.id ? (
          <Sidebar.Pusher
            as={StyledSegment}
            css={{
              maxWidth: '820px',
              paddingTop: 0,
              margin: '0 !important'
            }}
            basic
          >
            {selectedBook && selectedBook.id ? <Reader /> : null}
          </Sidebar.Pusher>
        ) : null}
      </Sidebar.Pushable>
    </div>
  );
};

export default connect(state => ({
  selectedBook: selectedBook(state),
  selectedPage: selectedPage(state)
}))(App);
