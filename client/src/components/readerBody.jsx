import React from 'react';

import { Loader } from 'semantic-ui-react';

import { connect } from 'react-redux';
import { selectedPage, pagesIsLoading } from '../store';

const ReaderBody = ({ page, isLoading, format }) => {
  return isLoading ? <Loader active inline="centered" /> : <p>{page.text}</p>;
};

export default connect(state => ({
  page: selectedPage(state),
  isLoading: pagesIsLoading(state)
}))(ReaderBody);
