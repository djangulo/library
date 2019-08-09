import React from 'react';

import { NavLink } from 'react-router-dom';

import { Item, Loader } from 'semantic-ui-react';

const BookList = ({
  isLoading,
  books,
  selectBook,
  selectedBook,
  lastPages,
  setPage,
  currentPaginationItem,
  location,
  match,
  searchItems,
  setSearchItems,
  setSearchQuery
}) => {
  return searchItems && searchItems.length > 0 ? (
    <Item.Group id="book-list" divided relaxed>
      {searchItems.map(b => (
        <Item
          key={b.id}
          className="book"
          as={NavLink}
          to={
            selectedBook.id === b.id
              ? '/books'
              : `/books/${b.slug}${
                  lastPages[b.id] && lastPages[b.id] !== 1
                    ? `?page=${lastPages[b.id]}`
                    : ''
                }`
          }
          onClick={() => {
            selectBook(b);
            setSearchItems([]);
            setSearchQuery('');
            // const qParams = new URLSearchParams(location.search);
            // const page = qParams.get('page');

            if (lastPages && lastPages[b.id]) {
              setPage(b.id, lastPages[b.id]);
            } else {
              setPage(b.id, 1);
            }
          }}
        >
          <Item.Content>
            <Item.Header>{b.title}</Item.Header>
            <Item.Meta>
              <span className="cinema">
                {b.page_count ? `${b.page_count} pages` : ''}
                {b.author ? ` | ${b.author}` : ''}
                {b.pub_year ? ` | ${b.pub_year}` : ''}
              </span>
            </Item.Meta>
            <Item.Description>{b.synopsis}</Item.Description>
          </Item.Content>
        </Item>
      ))}
    </Item.Group>
  ) : currentPaginationItem.data ? (
    <>
      <Item.Group id="book-list" divided relaxed>
        {currentPaginationItem.data.map(b => (
          <Item
            key={b.id}
            className="book"
            as={NavLink}
            to={
              selectedBook.id === b.id
                ? '/books'
                : `/books/${b.slug}${
                    lastPages[b.id] && lastPages[b.id] !== 1
                      ? `?page=${lastPages[b.id]}`
                      : ''
                  }`
            }
            onClick={() => {
              selectBook(b);
              // const qParams = new URLSearchParams(location.search);
              // const page = qParams.get('page');

              if (lastPages && lastPages[b.id]) {
                setPage(b.id, lastPages[b.id]);
              } else {
                setPage(b.id, 1);
              }
            }}
          >
            <Item.Content>
              <Item.Header>{b.title}</Item.Header>
              <Item.Meta>
                <span className="cinema">
                  {b.page_count ? `${b.page_count} pages` : ''}
                  {b.author ? ` | ${b.author}` : ''}
                  {b.pub_year ? ` | ${b.pub_year}` : ''}
                </span>
              </Item.Meta>
              <Item.Description>{b.synopsis}</Item.Description>
            </Item.Content>
          </Item>
        ))}
      </Item.Group>
      {isLoading ? <Loader active inline="centered" /> : null}
    </>
  ) : null;
};

export default BookList;
