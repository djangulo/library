import React from 'react';
import PropTypes from 'prop-types';

const List = ({ id, className, children }) => {
  return (
    <div id={id} className={className}>
      {children}
    </div>
  );
};

List.propTypes = {
  id: PropTypes.string,
  className: PropTypes.string,
  children: PropTypes.arrayOf(PropTypes.element)
};

export default List;
