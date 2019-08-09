const paginate = require('./paginate');
const paginateCount = require('../config').paginateBy;
const rootUrl = require('../config').rootUrl;

const resolveNext = (url, total, current) => {
  if (current === total) return null;
  return url + '?page=' + (current + 1);
};

const resolvePrev = (url, total, current) => {
  if (current === 1) return null;
  return url + (current - 1 === 1 ? '' : '?page=' + (current - 1));
};

const paginatedResponse = (count, path, page, items) => {
  if (count === 0 || parseInt(count, 10) === 0)
    return {
      items: 0,
      pages: 0,
      current: 0,
      previous: null,
      next: null,
      data: []
    };
  const baseUrl = rootUrl + path;
  const total = Math.ceil(items.length / paginateCount);
  if (page > total) {
    return {
      items: count,
      pages: total,
      current: total,
      previous: resolvePrev(baseUrl, total, total),
      next: null,
      data: paginate(items, total, paginateCount)
    };
  }
  return {
    items: count,
    pages: total,
    current: page,
    previous: resolvePrev(baseUrl, total, page),
    next: resolveNext(baseUrl, total, page),
    data: paginate(items, page, paginateCount)
  };
};

module.exports = paginatedResponse;
