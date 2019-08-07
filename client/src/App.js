/** @jsx jsx */
import React from "react";
import { jsx } from "@emotion/core";
import styled from "@emotion/styled";

import { connect } from "react-redux";

import { Route } from "react-router-dom";

import { Helmet } from "react-helmet";

import BookList from "./components/bookList";

import { Sidebar, Header, Segment } from "semantic-ui-react";
import { selectedBook, selectedPage } from "./store";
import Reader from "./components/reader";

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

const App = ({ selectedBook, selectedPage }) => {
  const [visible] = React.useState(true);
  return (
    <div className='App'>
      <Helmet>
        <title>Library</title>
      </Helmet>
      <HeaderComponent className='App-header'>
        <Header as='h2' floated='left'>
          <Header.Content as='a' href='/'>
            Library
          </Header.Content>
        </Header>
      </HeaderComponent>
      <Sidebar.Pushable as={StyledSegment} css={{ marginTop: 0 }}>
        <Sidebar
          as={Segment}
          animation='slide out'
          icon='labeled'
          vertical
          visible={visible}
          width='wide'
        >
          <BookList />
        </Sidebar>

        {selectedBook && selectedBook.id ? (
          <Route path='/books/:slug'>
            <Sidebar.Pusher
              as={StyledSegment}
              css={{
                maxWidth: "820px",
                paddingTop: 0,
                margin: "0 !important"
              }}
              basic
            >
              {selectedBook && selectedBook.id ? <Reader /> : null}
            </Sidebar.Pusher>
          </Route>
        ) : null}
      </Sidebar.Pushable>
    </div>
  );
};

export default connect(state => ({
  selectedBook: selectedBook(state),
  selectedPage: selectedPage(state)
}))(App);
