const errors = require('../../../client/src/data/errors');

describe('Juan visits the library', () => {
  // Juan just heard of this new library app, he decides to check it out
  it('successfully loads', () => {
    cy.visit('/');
  });
  // He immediately sees a list of many different books
  it('should display a list of books', () => {
    cy.get('#book-list').should('exist');
  });
  //  He clicks on one of them, and notices a reading pane opens up
  it('should open a reading panel when clicking on a book', () => {
    cy.get('.book')
      .contains('It')
      .click();
    cy.get('#reader').should('exist');
  });

  // He then notices the total pages displayed  in there too
  it('should display the page count', () => {
    cy.get('span')
      .contains('Page 1 of 20')
      .should('exist');
  });

  // Juan notices there is a curious input with a 'jump to' placeholder,
  // he decides to test the developer by entering some bogus values
  // He tries to click previous page, but gets an error because he is already
  // on page 1
  it('should yield an error when page cannot be decreased further', () => {
    cy.get('.button')
      .contains('Previous')
      .click();
    cy.get('.error > p').should('contain', errors.positivePageNumber);
  });
  it('should validate against non-numeric fields', () => {
    cy.get('input[name="jump-to"]').type('abcdef{enter}');
    cy.get('.error > p').should('contain', errors.mustBeNumeric);
  });
  it('should validate against negative numbers', () => {
    cy.get('input[name="jump-to"]').type('-1{enter}');
    cy.get('.error > p').should('contain', errors.positivePageNumber);
  });
  it('should validate against numbers higher than page count', () => {
    cy.get('input[name="jump-to"]').type('20000{enter}');
    cy.get('.error > p').should('contain', errors.cannotExceedPageCount(20));
  });

  it('should hide when the selected book is clicked', () => {
    cy.get('.book')
      .contains('It')
      .click();
    cy.get('#reader').should('not.exist');
  });
});
