/** @jsx jsx */
// eslint-disable-next-line
import React from 'react';
import { jsx } from '@emotion/core';
import { Loader, Container } from 'semantic-ui-react';

import { connect } from 'react-redux';
import { selectedPage, pagesIsLoading } from '../store';

const ReaderBody = ({ page, isLoading }) => {
  return isLoading ? (
    <Loader active inline="centered" />
  ) : (
    <Container fluid css={{ overflowY: 'visible' }}>
      <p css={{ whiteSpace: 'pre-wrap' }}>
        {page && page.body ? page.body : null}
      </p>
    </Container>
  );
};

export default connect(state => ({
  page: selectedPage(state),
  isLoading: pagesIsLoading(state)
}))(ReaderBody);
