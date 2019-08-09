function getApiRoot() {
  switch (process.env.NODE_ENV) {
    case 'production':
      return 'https://library-api.djangulo.com';
    default:
      return 'http://localhost:9000';
  }
}

export default {
  apiRoot: getApiRoot()
};
