import { combineReducers } from 'redux';

import books from './booksDuck';
import pages from './pagesDuck';
import search from './searchDuck';

export * from './booksDuck';
export * from './pagesDuck';
export * from './searchDuck';

export const rootReducer = combineReducers({
  books,
  pages,
  search
});
