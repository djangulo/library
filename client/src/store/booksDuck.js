// QUACK! This is a duck. https://github.com/erikras/ducks-modular-redux
import uniqBy from 'lodash/uniqBy';
import orderBy from 'lodash/orderBy';
import bookService from '../services/bookService';

// sync actions
const SELECT_BOOK = 'SELECT_BOOK';
const SORT_BY_COLUMN = 'SORT_BY_COLUMN';
const SELECT_PAGINATION_ITEM = 'SELECT_PAGINATION_ITEM';
const SET_SEARCH_ITEMS = 'SET_SEARCH_ITEMS';
const SET_SEARCH_QUERY = 'SET_SEARCH_QUERY';

// async actions
const REQUEST_BOOKS = 'REQUEST_BOOKS';
const REQUEST_BOOKS_SUCCESS = 'REQUEST_BOOKS_SUCCESS';
const REQUEST_BOOKS_FAILURE = 'REQUEST_BOOKS_FAILURE';

const SEARCH_BOOKS = 'SEARCH_BOOKS';
const SEARCH_BOOKS_SUCCESS = 'SEARCH_BOOKS_SUCCESS';
const SEARCH_BOOKS_FAILURE = 'SEARCH_BOOKS_FAILURE';

// pages = {
//   1: {
//     items: 0,
//     pages: 10,
//     previous: null,
//     next: null,
//     data: null
//   }
// };

const initialState = {
  items: [],
  searchItems: [],
  searchQuery: '',
  page: 1,
  currentPage: null,
  pages: {},
  selected: undefined,
  error: null,
  isLoading: false,
  sortColumn: 'title',
  sortDirection: 'asc'
};

// action creators
export const selectBook = book => ({ type: SELECT_BOOK, book });
export const setSearchQuery = query => ({ type: SET_SEARCH_QUERY, query });
export const selectPaginationItem = pageNum => ({
  type: SELECT_PAGINATION_ITEM,
  pageNum
});
export const sortByColumn = column => ({
  type: SORT_BY_COLUMN,
  column
});
export const requestBooks = (pageNum = 1) => ({ type: REQUEST_BOOKS, pageNum });
export const requestBooksFailure = error => ({
  type: REQUEST_BOOKS_FAILURE,
  error
});
export const requestBooksSuccess = json => ({
  type: REQUEST_BOOKS_SUCCESS,
  json
});

export const searchRequest = (query, sort = 'title', order = 'asc') => ({
  type: SEARCH_BOOKS,
  query,
  sort,
  order
});
export const searchRequestFailure = error => ({
  type: SEARCH_BOOKS_FAILURE,
  error
});
export const searchRequestSuccess = json => ({
  type: SEARCH_BOOKS_SUCCESS,
  json
});

export const setSearchItems = items => ({
  type: SET_SEARCH_ITEMS,
  items
});
export const searchBooks = (q, sort = 'title', order = 'asc') => (
  dispatch,
  getState
) => {
  dispatch(searchRequest(q, sort, order));
  const allItems = getState().books.items;
  const totalItems = getState().books.pages[1].items;
  // append local items first
  const filtered = getState().books.items.filter(
    b =>
      b.title.toLowerCase().includes(q.toLowerCase()) ||
      (b.author && b.author.toLowerCase().includes(q.toLowerCase()))
  );
  const ordered = orderBy(filtered, [sort], [order]);
  dispatch(setSearchItems(ordered));
  // search the server now
  // only if there are actually more items to search
  if (allItems < totalItems) {
    return bookService
      .search(q, sort, order)
      .then(
        response => response.json(),
        error => dispatch(requestBooksFailure(error))
      )
      .then(json => {
        if (json && json.data) dispatch(searchRequestSuccess(json));
      });
  } else {
    //don't hit the server at all
    searchRequestSuccess({ data: [] });
    return;
  }
};

export const fetchBooks = (pageNum = 1) => (dispatch, getState) => {
  dispatch(requestBooks(pageNum));
  const isHere = getState().books.pages[pageNum] ? true : false;
  if (isHere) {
    dispatch(selectPaginationItem(pageNum));
    return;
  }
  return bookService
    .list(pageNum)
    .then(
      response => response.json(),
      error => dispatch(requestBooksFailure(error))
    )
    .then(json => {
      dispatch(requestBooksSuccess(json));
      dispatch(selectPaginationItem(pageNum));
      return json;
    });
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
          state.selected && action.book && action.book.id === state.selected.id
            ? null
            : action.book
      };
    case SELECT_PAGINATION_ITEM:
      return {
        ...state,
        page: action.pageNum,
        currentPage: state.pages[action.pageNum],
        isLoading: false
      };
    case SORT_BY_COLUMN:
      return {
        ...state,
        sortColumn: action.column,
        sortDirection: resolveSortOrder(state, action.column),
        currentPage: {
          ...state.currentPage,
          data: orderBy(
            state.pages[state.page].data,
            [action.column],
            [resolveSortOrder(state, action.column)]
          )
        },
        pages: {
          ...state.pages,
          [state.page]: {
            ...state.pages[state.page],
            data: orderBy(
              state.pages[state.page].data,
              [action.column],
              [resolveSortOrder(state, action.column)]
            )
          }
        }
      };
    case SEARCH_BOOKS:
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
        items: uniqBy([...state.items, ...action.json.data], 'id'),
        pages: { ...state.pages, [action.json.current]: action.json }
      };
    case SEARCH_BOOKS_SUCCESS:
      return {
        ...state,
        isLoading: false,
        searchItems: uniqBy([...state.searchItems, ...action.json.data], 'id')
      };
    case SET_SEARCH_ITEMS:
      return {
        ...state,
        searchItems: [...action.items]
      };
    case REQUEST_BOOKS_FAILURE:
    case SEARCH_BOOKS_FAILURE:
      return {
        ...state,
        error: action.error
      };
    case SET_SEARCH_QUERY:
      return {
        ...state,
        searchQuery: action.query
      };
    default:
      return state;
  }
};
export default reducer;

// selectors
export const getBooks = ({ books: { items } }) => items || [];
export const selectedBook = ({ books: { selected } }) => selected || {};
export const selectedPaginationItem = ({ books: { currentPage } }) =>
  currentPage || {};
export const getPaginationItems = ({ books: { pages } }) => pages || {};
export const booksError = ({ books: { error } }) => error || '';
export const booksIsLoading = ({ books: { isLoading } }) => isLoading || null;
export const booksSortColumn = ({ books: { sortColumn } }) =>
  sortColumn || 'title';
export const booksSortDirection = ({ books: { sortDirection } }) =>
  sortDirection || 'asc';

export const searchQuery = ({ books: { searchQuery } }) => searchQuery || '';
export const searchItems = ({ books: { searchItems } }) => searchItems || [];
