const positivePageNumber = 'Page numbers must be positive.';
const cannotExceedPageCount = pageCount =>
  `Cannot exceed page count (${pageCount}).`;
const mustBeNumeric = 'Page numbers must be numeric.';

export default {
  positivePageNumber,
  cannotExceedPageCount,
  mustBeNumeric
};
