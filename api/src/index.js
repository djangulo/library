const express = require('express');

/**
 * @typedef AppRoutes
 * @type {Object}
 * @property {string} path path to assign to the route; e.g. /books
 * @property {express.Router} router router object
 * @param {function[]} middleware Array of middleware to apply
 * @param {AppRoutes} appRoutes routes to implement
 * @param {Object} context context object to pass to the requests
 */
const createServer = async (middleware = [], appRoutes = [], context = {}) => {
  const app = await express();
  for (let m of middleware) {
    await app.use(m());
  }
  for (let r of appRoutes) {
    await app.use(r.path, r.router);
  }
  app.use((req, res, next) => {
    req.context = context;
    next();
  });
  return app;
};

module.exports = createServer;
