import { combineReducers } from 'redux';

import books from './booksDuck';
import pages from './pagesDuck';

export * from './booksDuck';
export * from './pagesDuck';

export const rootReducer = combineReducers({
  books,
  pages
});
