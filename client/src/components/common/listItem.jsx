import React from 'react';
import PropTypes from 'prop-types';

const ListItem = ({ text, subtext }) => {
  return (
    <div className="list-item">
      <p className="list-item-text">{text}</p>
      <p className="list-item-subtext">{subtext}</p>
    </div>
  );
};

ListItem.propTypes = {
  text: PropTypes.string,
  subtext: PropTypes.string
};

export default ListItem;
