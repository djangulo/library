import React from 'react';
import { connect, useDispatch } from 'react-redux';

import { Button, Form, Header, Item, Loader } from 'semantic-ui-react';

import {
  getBooks,
  selectedBook,
  booksError,
  booksIsLoading,
  booksSortColumn,
  selectBook,
  fetchBooks,
  setPage,
  booksSortDirection,
  sortByColumn,
  lastPageByBook
} from '../store';

const BookList = ({
  books,
  selectedBook,
  error,
  isLoading,
  fetchBooks,
  selectBook,
  sortColumn,
  sortDirection,
  sortByColumn,
  lastPages,
  setPage
}) => {
  const dispatch = useDispatch();
  React.useEffect(() => {
    fetchBooks();
  }, [dispatch, fetchBooks]);

  const sortOptions = [
    { key: 'title', value: 'title', text: 'Title' },
    { key: 'author', value: 'author', text: 'Author' },
    {
      key: 'publication_date',
      value: 'publication_date',
      text: 'Pub. Date'
    },
    { key: 'page_count', value: 'page_count', text: 'Pages' }
  ];

  const resolveIcon = column => {
    if (sortColumn !== column) return null;
    if (sortDirection === 'asc') return 'triangle down';
    if (sortDirection === 'desc') return 'triangle up';
  };

  return (
    <>
      <div>
        <Header as="h3">
          <Header.Content>Listings</Header.Content>
        </Header>
        <Form.Group>
          <Form.Field>
            <label>Sort by</label>
            <Button.Group>
              {sortOptions.map(o => (
                <Button
                  compact
                  size="mini"
                  key={o.key}
                  onClick={() => sortByColumn(o.value)}
                  icon={resolveIcon(o.value)}
                  content={o.text}
                />
              ))}
            </Button.Group>
          </Form.Field>
        </Form.Group>
      </div>
      {error ? <span>{error}</span> : null}
      <Item.Group id="book-list" divided relaxed>
        {books.map(b => (
          <Item
            key={b.id}
            className="book"
            as="a"
            onClick={() => {
              selectBook(b);
              setPage(b.id, lastPages[b.id] ? lastPages[b.id] : 1);
            }}
          >
            <Item.Content>
              <Item.Header>{b.title}</Item.Header>
              <Item.Meta>
                <span className="cinema">
                  {b.author} | {b.publication_date} | {b.page_count} pages
                </span>
              </Item.Meta>
              <Item.Description>{b.synopsis}</Item.Description>
            </Item.Content>
          </Item>
        ))}
      </Item.Group>
      {isLoading ? <Loader active inline="centered" /> : null}
    </>
  );
};

export default connect(
  state => ({
    books: getBooks(state),
    selectedBook: selectedBook(state),
    error: booksError(state),
    isLoading: booksIsLoading(state),
    sortColumn: booksSortColumn(state),
    sortDirection: booksSortDirection(state),
    lastPages: lastPageByBook(state)
  }),
  { fetchBooks, selectBook, setPage, sortByColumn }
)(BookList);
