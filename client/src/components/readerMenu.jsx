import React from 'react';

import { Input, Button, Menu } from 'semantic-ui-react';
import { connect } from 'react-redux';
import { setPage, selectedPage, selectedBook, setPagesError } from '../store';

import errors from '../data/errors';

const ReaderMenu = ({ book, page, setPage, setError }) => {
  const [val, setVal] = React.useState('');

  const handlePreviousPage = () => {
    if (page.number > 1) {
      setPage(book.id, page.number - 1);
      setError(null);
    } else {
      setError(errors.positivePageNumber);
    }
  };
  const handleNextPage = () => {
    if (page.number < book.page_count) {
      setPage(book.id, page.number + 1);
      setError(null);
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
        <Button compact onClick={() => handlePreviousPage()}>
          Previous
        </Button>
      </Menu.Item>
      <Menu.Item>
        <span>
          Page {page.number} of {book.page_count}
        </span>
      </Menu.Item>
      <Menu.Item>
        <Button compact onClick={() => handleNextPage()}>
          Next
        </Button>
      </Menu.Item>
      <Menu.Item position="right">
        <Input
          name="jump-to"
          compact
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

export default connect(
  state => ({
    book: selectedBook(state),
    page: selectedPage(state)
  }),
  { setPage, setError: setPagesError }
)(ReaderMenu);
