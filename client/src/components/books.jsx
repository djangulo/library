import React from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router-dom';

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
  lastPageByBook,
  selectPaginationItem,
  getPaginationItems,
  selectedPaginationItem,
  searchQuery,
  setSearchQuery,
  searchBooks,
  searchItems,
  setSearchItems
} from '../store';
import BookList from './bookList';
import BookHeader from './bookHeader';

const Books = ({
  books,
  selectedBook,
  error,
  isLoading,
  fetchBooks,
  selectBook,
  sortColumn,
  sortDirection,
  sortByColumn,
  selectPaginationItem,
  paginationItems,
  currentPaginationItem,
  lastPages,
  setPage,
  history,
  match,
  location,
  searchQuery,
  setSearchQuery,
  searchBooks,
  searchItems,
  setSearchItems
}) => {
  return (
    <>
      <BookHeader
        sortColumn={sortColumn}
        sortDirection={sortDirection}
        sortByColumn={sortByColumn}
        error={error}
        location={location}
        paginationItems={paginationItems}
        selectPaginationItem={selectPaginationItem}
        currentPaginationItem={currentPaginationItem}
        fetchBooks={fetchBooks}
        searchQuery={searchQuery}
        setSearchQuery={setSearchQuery}
        setSearchItems={setSearchItems}
        searchBooks={searchBooks}
      />
      <BookList
        books={books}
        currentPaginationItem={currentPaginationItem}
        selectedBook={selectedBook}
        isLoading={isLoading}
        lastPages={lastPages}
        selectBook={selectBook}
        setPage={setPage}
        match={match}
        location={location}
        setSearchQuery={setSearchQuery}
        searchItems={searchItems}
        setSearchItems={setSearchItems}
      />
    </>
  );
};

export default withRouter(
  connect(
    state => ({
      books: getBooks(state),
      selectedBook: selectedBook(state),
      error: booksError(state),
      isLoading: booksIsLoading(state),
      sortColumn: booksSortColumn(state),
      sortDirection: booksSortDirection(state),
      lastPages: lastPageByBook(state),
      paginationItems: getPaginationItems(state),
      currentPaginationItem: selectedPaginationItem(state),
      searchQuery: searchQuery(state),
      searchItems: searchItems(state)
    }),
    {
      fetchBooks,
      selectBook,
      setPage,
      sortByColumn,
      selectPaginationItem,
      setSearchQuery,
      searchBooks,
      setSearchItems
    }
  )(Books)
);
