import React from 'react';
import { connect } from 'react-redux';

import { NavLink } from 'react-router-dom';

import { Item, Loader } from 'semantic-ui-react';
import {
  selectedBook,
  booksIsLoading,
  selectBook,
  setPage,
  lastPageByBook,
  selectedPaginationItem,
  setSearchQuery,
  searchItems,
  setSearchItems,
  searchQuery
} from '../store';

const BookList = ({
  isLoading,
  selectBook,
  selectedBook,
  lastPages,
  setPage,
  currentPaginationItem,
  searchItems,
  setSearchItems,
  setSearchQuery,
  searchQuery
}) => {
  return searchQuery.length > 0 ? (
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
            setSearchQuery('');
            selectBook(b);
            setSearchItems([]);

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

export default connect(
  state => ({
    isLoading: booksIsLoading(state),
    selectedBook: selectedBook(state),
    lastPages: lastPageByBook(state),
    currentPaginationItem: selectedPaginationItem(state),
    searchItems: searchItems(state),
    searchQuery: searchQuery(state)
  }),
  { selectBook, setPage, setSearchItems, setSearchQuery }
)(BookList);
