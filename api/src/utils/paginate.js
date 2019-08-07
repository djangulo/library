const paginate = (items, pageNumber, pageSize) => {
  var totalPages = Math.ceil(items.length / pageSize);
  var idx = (pageNumber - 1) * pageSize;
  if (pageNumber > totalPages) {
    return paginate(items, totalPages, pageSize);
  }
  if (items.length % pageSize !== 0 && pageNumber > totalPages) {
    return items.slice(idx);
  }
  return items.slice(idx, pageNumber * pageSize);
};

module.exports = { paginate };
