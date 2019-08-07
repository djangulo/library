function getApiRoot() {
  console.log(process.env.NODE_ENV);
  switch (process.env.NODE_ENV) {
    case 'production':
      return 'https://library-api-staging.djangulo.com';
    default:
      return 'http://localhost:9000';
  }
}

export default {
  apiRoot: getApiRoot()
};
