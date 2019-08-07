import React from "react";
import { connect, useDispatch } from "react-redux";

import { NavLink } from "react-router-dom";

import { Button, Form, Header, Item, Loader } from "semantic-ui-react";

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
} from "../store";

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
  setPage,
  history
}) => {
  const dispatch = useDispatch();
  React.useEffect(() => {
    fetchBooks();
  }, [dispatch, fetchBooks]);

  const sortOptions = [
    { key: "title", value: "title", text: "Title" },
    { key: "author", value: "author", text: "Author" },
    {
      key: "pub_year",
      value: "pub_year",
      text: "Pub. Date"
    },
    { key: "page_count", value: "page_count", text: "Pages" }
  ];

  const resolveIcon = column => {
    if (sortColumn !== column) return null;
    if (sortDirection === "asc") return "triangle down";
    if (sortDirection === "desc") return "triangle up";
  };

  return (
    <>
      <div>
        <Header as='h3'>
          <Header.Content>Listings</Header.Content>
        </Header>
        <Form.Group>
          <Form.Field>
            <label>Sort by</label>
            <Button.Group>
              {sortOptions.map(o => (
                <Button
                  compact
                  size='mini'
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
      {error ? <span>{error.toString()}</span> : null}
      <Item.Group id='book-list' divided relaxed>
        {books.map(b => (
          <Item
            key={b.id}
            className='book'
            as={NavLink}
            to={
              selectedBook.id === b.id
                ? "/books"
                : `/books/${b.slug}${
                    lastPages[b.id] && lastPages[b.id] !== 1
                      ? `?page=${lastPages[b.id]}`
                      : ""
                  }`
            }
            onClick={() => {
              selectBook(b);
              setPage(b.id, lastPages[b.id] ? lastPages[b.id] : 1);
            }}
          >
            <Item.Content>
              <Item.Header>{b.title}</Item.Header>
              <Item.Meta>
                <span className='cinema'>
                  {b.page_count ? `${b.page_count} pages` : ""}
                  {b.author ? ` | ${b.author}` : ""}
                  {b.pub_year ? ` | ${b.pub_year}` : ""}
                </span>
              </Item.Meta>
              <Item.Description>{b.synopsis}</Item.Description>
            </Item.Content>
          </Item>
        ))}
      </Item.Group>
      {isLoading ? <Loader active inline='centered' /> : null}
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
