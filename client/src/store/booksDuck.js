// QUACK! This is a duck. https://github.com/erikras/ducks-modular-redux
import uniqBy from 'lodash/uniqBy';
import orderBy from 'lodash/orderBy';
import bookService from '../services/fakeBookService';

// sync actions
const SELECT_BOOK = 'SELECT_BOOK';
const SORT_BY_COLUMN = 'SORT_BY_COLUMN';

// async actions
const REQUEST_BOOKS = 'REQUEST_BOOKS';
const REQUEST_BOOKS_SUCCESS = 'REQUEST_BOOKS_SUCCESS';
const REQUEST_BOOKS_FAILURE = 'REQUEST_BOOKS_FAILURE';

const initialState = {
  items: [],
  selected: null,
  error: null,
  isLoading: false,
  sortColumn: 'title',
  sortDirection: 'asc'
};

// action creators
export const selectBook = book => ({ type: SELECT_BOOK, book });
export const sortByColumn = column => ({
  type: SORT_BY_COLUMN,
  column
});
export const requestBooks = () => ({ type: REQUEST_BOOKS });
export const requestBooksFailure = error => ({
  type: REQUEST_BOOKS_FAILURE,
  error
});
export const requestBooksSuccess = json => ({
  type: REQUEST_BOOKS_SUCCESS,
  json
});

export const fetchBooks = () => dispatch => {
  dispatch(requestBooks());
  return bookService
    .list()
    .then(response => response, error => dispatch(requestBooksFailure(error)))
    .then(json => dispatch(requestBooksSuccess(json)));
};

const resolveSortOrder = (state, column) => {
  const { sortColumn, sortDirection } = state;
  if (sortColumn === column) {
    return sortDirection === 'asc' ? 'desc' : 'asc';
  }
  return 'asc';
};

// Reducer
const reducer = (state = initialState, action = {}) => {
  switch (action.type) {
    case SELECT_BOOK:
      return {
        ...state,
        selected:
          state.selected && action.book.id === state.selected.id
            ? null
            : action.book
      };
    case SORT_BY_COLUMN:
      return {
        ...state,
        sortColumn: action.column,
        sortDirection: resolveSortOrder(state, action.column),
        items: orderBy(
          [...state.items],
          [action.column],
          [resolveSortOrder(state, action.column)]
        )
      };
    case REQUEST_BOOKS:
      return {
        ...state,
        isLoading: true,
        error: null
      };
    case REQUEST_BOOKS_SUCCESS:
      return {
        ...state,
        isLoading: false,
        items: uniqBy([...state.items, ...action.json], 'id')
      };
    case REQUEST_BOOKS_FAILURE:
      return {
        ...state,
        error: action.error
      };
    default:
      return state;
  }
};
export default reducer;

// selectors
export const getBooks = ({ books: { items } }) => items || [];
export const selectedBook = ({ books: { selected } }) => selected || {};
export const booksError = ({ books: { error } }) => error || '';
export const booksIsLoading = ({ books: { isLoading } }) => isLoading || null;
export const booksSortColumn = ({ books: { sortColumn } }) =>
  sortColumn || 'title';
export const booksSortDirection = ({ books: { sortDirection } }) =>
  sortDirection || 'asc';
