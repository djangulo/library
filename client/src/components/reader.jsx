/** @jsx jsx */
// eslint-disable-next-line
import React from "react";
import { jsx } from "@emotion/core";
import ReaderHeader from "./readerHeader";
import ReaderBody from "./readerBody";

import { Helmet } from "react-helmet";

import { connect } from "react-redux";
import { selectedBook } from "./../store/booksDuck";

const Reader = ({ book }) => {
  return (
    <div id='reader' css={{ margin: "auto" }}>
      <Helmet>
        <title>
          Library
          {book && book.title
            ? ` | Read ${book.title} by ${book.author}`
            : null}
        </title>
      </Helmet>
      <ReaderHeader />
      <ReaderBody />
    </div>
  );
};

export default connect(state => ({
  book: selectedBook(state)
}))(Reader);
