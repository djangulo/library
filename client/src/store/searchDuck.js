// QUACK! This is a duck. https://github.com/erikras/ducks-modular-redux
import uniqBy from 'lodash/uniqBy';
import orderBy from 'lodash/orderBy';
import bookService from '../services/bookService';

// sync actions
const SET_SEARCH_ITEMS = 'SET_SEARCH_ITEMS';
const SET_SEARCH_QUERY = 'SET_SEARCH_QUERY';

// async actions
const SEARCH_BOOKS = 'SEARCH_BOOKS';
const SEARCH_BOOKS_SUCCESS = 'SEARCH_BOOKS_SUCCESS';
const SEARCH_BOOKS_FAILURE = 'SEARCH_BOOKS_FAILURE';

const initialState = {
  items: [],
  query: '',
  error: null,
  isLoading: false
};

// action creators
export const setSearchQuery = query => ({ type: SET_SEARCH_QUERY, query });
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
  // append local items first
  const filtered = getState().books.items.filter(
    b =>
      b.title.toLowerCase().includes(q.toLowerCase()) ||
      (b.author && b.author.toLowerCase().includes(q.toLowerCase()))
  );
  const ordered = orderBy(filtered, [sort], [order]);
  dispatch(setSearchItems(ordered));
  // search the server now
  return bookService
    .search(q, sort, order)
    .then(
      response => response.json(),
      error => dispatch(searchRequestFailure(error))
    )
    .then(json => {
      if (json && json.data) dispatch(searchRequestSuccess(json));
    });
};

// Reducer
const reducer = (state = initialState, action = {}) => {
  switch (action.type) {
    case SEARCH_BOOKS:
      return {
        ...state,
        isLoading: true,
        error: null
      };
    case SEARCH_BOOKS_SUCCESS:
      return {
        ...state,
        isLoading: false,
        items: uniqBy([...state.items, ...action.json.data], 'id')
      };
    case SET_SEARCH_ITEMS:
      return {
        ...state,
        items: [...action.items]
      };
    case SEARCH_BOOKS_FAILURE:
      return {
        ...state,
        error: action.error
      };
    case SET_SEARCH_QUERY:
      return {
        ...state,
        query: action.query
      };
    default:
      return state;
  }
};
export default reducer;

// selectors
export const searchQuery = ({ search: { query } }) => query || '';
export const searchItems = ({ search: { items } }) => items || [];
