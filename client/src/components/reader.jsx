/** @jsx jsx */
// eslint-disable-next-line
import React from 'react';
import { jsx } from '@emotion/core';
import ReaderHeader from './readerHeader';
import ReaderBody from './readerBody';

const Reader = () => {
  return (
    <div id="reader" css={{ margin: 'auto' }}>
      <ReaderHeader />
      <ReaderBody />
    </div>
  );
};

export default Reader;
