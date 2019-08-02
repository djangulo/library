// QUACK! This is a duck. https://github.com/erikras/ducks-modular-redux
import uniqBy from 'lodash/uniqBy';
import pageService from '../services/fakePageService';

// sync actions
const SELECT_PAGE = 'SELECT_PAGE';
const SET_ERROR = 'SET_ERROR';

// async actions
const REQUEST_PAGE = 'REQUEST_PAGE';
const REQUEST_PAGE_SUCCESS = 'REQUEST_PAGE_SUCCESS';
const REQUEST_PAGE_FAILURE = 'REQUEST_PAGE_FAILURE';

const initialState = {
  items: [],
  selected: null,
  error: null,
  isLoading: false,
  lastPageByBook: {}
};

// action creators
export const selectPage = page => ({
  type: SELECT_PAGE,
  page
});
export const setPagesError = error => ({
  type: SET_ERROR,
  error
});
export const requestPage = () => ({ type: REQUEST_PAGE });
export const requestPageFailure = error => ({
  type: REQUEST_PAGE_FAILURE,
  error
});
export const requestPageSuccess = json => ({
  type: REQUEST_PAGE_SUCCESS,
  json
});

export const setPage = (bookId, pageNumber) => (dispatch, getState) => {
  const currentPage = { ...getState().pages.selected };
  const page = getState().pages.items.find(
    p => p.book === bookId && p.number === pageNumber
  );
  if (page) {
    dispatch(selectPage(page));
    return;
  }
  dispatch(requestPage());
  return pageService
    .retrieve(bookId, pageNumber)
    .then(
      response => response,
      error => {
        dispatch(requestPageFailure(error));
        dispatch(selectPage(currentPage));
      }
    )
    .then(json => {
      dispatch(requestPageSuccess(json));
      dispatch(selectPage(json));
    });
};

// Reducer
const reducer = (state = initialState, action = {}) => {
  switch (action.type) {
    case SELECT_PAGE:
      return {
        ...state,
        selected: action.page,
        lastPageByBook: {
          ...state.lastPageByBook,
          [action.page.book]: action.page.number
        }
      };
    case SET_ERROR:
      return {
        ...state,
        error: action.error
      };
    case REQUEST_PAGE:
      return {
        ...state,
        isLoading: true,
        error: null
      };
    case REQUEST_PAGE_SUCCESS:
      return {
        ...state,
        isLoading: false,
        items: uniqBy([...state.items, action.json], 'id') // single page with 'retrieve'
      };
    case REQUEST_PAGE_FAILURE:
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
export const getPages = ({ pages: { items } }) => items || [];
export const selectedPage = ({ pages: { selected } }) => selected || {};
export const pagesError = ({ pages: { error } }) => error || '';
export const pagesIsLoading = ({ pages: { isLoading } }) => isLoading || null;
export const lastPageByBook = ({ pages: { lastPageByBook } }) =>
  lastPageByBook || {};
