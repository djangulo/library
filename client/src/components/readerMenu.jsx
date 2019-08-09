import React from 'react';

import { withRouter } from 'react-router-dom';

import { Input, Button, Menu } from 'semantic-ui-react';
import { connect } from 'react-redux';
import { setPage, selectedPage, selectedBook, setPagesError } from '../store';

import errors from '../data/errors';

const ReaderMenu = ({ book, page, setPage, setError, history }) => {
  const [val, setVal] = React.useState('');

  const handlePreviousPage = () => {
    if (page.page_number > 1) {
      const newPageNumber = page.page_number - 1;
      setPage(book.id, newPageNumber);
      setError(null);
      history.push(
        `/books/${book.slug}${
          newPageNumber !== 1 ? `?page=${newPageNumber}` : ''
        }`
      );
    } else {
      setError(errors.positivePageNumber);
    }
  };
  const handleNextPage = () => {
    const newPageNumber = page.page_number + 1;
    if (page.page_number < book.page_count) {
      setPage(book.id, newPageNumber);
      setError(null);
      history.push(`/books/${book.slug}?page=${newPageNumber}`);
    } else {
      setError(errors.cannotExceedPageCount(book.page_count));
    }
  };

  const handleSetPage = pageNumber => {
    if (isNaN(pageNumber)) {
      setError(errors.mustBeNumeric);
      setVal('');
      return;
    }
    const pgNum = parseInt(pageNumber, 10);
    if (pgNum >= 1 && pgNum <= book.page_count) {
      setPage(book.id, pgNum);
      setVal('');
      setError(null);
      history.push(`/books/${book.slug}${pgNum !== 1 ? `?page=${pgNum}` : ''}`);
    } else if (pgNum < 1) {
      setError(errors.positivePageNumber);
      setVal('');
    } else if (pgNum > book.page_count) {
      setError(errors.cannotExceedPageCount(book.page_count));
      setVal('');
    }
  };
  return (
    <Menu borderless secondary>
      <Menu.Item>
        <Button
          disabled={page.page_number === 1}
          compact
          onClick={() => handlePreviousPage()}
        >
          Previous
        </Button>
      </Menu.Item>
      <Menu.Item>
        <span>
          Page {page.page_number} of {book.page_count}
        </span>
      </Menu.Item>
      <Menu.Item>
        <Button
          disabled={page.page_number === book.page_count}
          compact
          onClick={() => handleNextPage()}
        >
          Next
        </Button>
      </Menu.Item>
      <Menu.Item position="right">
        <Input
          name="jump-to"
          value={val}
          onChange={(e, d) => setVal(d.value)}
          placeholder="Jump to"
          onKeyDown={e => {
            if (e.key === 'Enter') {
              handleSetPage(val);
            }
          }}
          action={{
            content: 'Go',
            onClick: () => handleSetPage(val)
          }}
        />
      </Menu.Item>
    </Menu>
  );
};

export default withRouter(
  connect(
    state => ({
      book: selectedBook(state),
      page: selectedPage(state)
    }),
    { setPage, setError: setPagesError }
  )(ReaderMenu)
);
