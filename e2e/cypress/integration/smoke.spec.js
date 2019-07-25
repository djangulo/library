/// <reference types="Cypress" />
describe('smoke test', () => {
  it('Checks basic math', () => {
    expect(2 + 2).to.equal(4);
  });
});
