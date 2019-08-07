module.exports = {
  servePort: parseInt(process.env.SERVE_PORT, 10) || 9000,
  paginate: parseInt(process.env.PAGINATE, 10) === 1 ? true : false || true,
  paginateBy: parseInt(process.env.PAGINATE_BY, 10) || 10,
  paragraphsPerPage: parseInt(process.env.PARAGRAPHS_PER_PAGE, 10) || 50,
  dbUser: process.env.POSTGRES_USER || 'postgres',
  dbPass: process.env.POSTGRES_PASSWORD || '',
  dbHost: process.env.POSTGRES_HOST || 'localhost',
  dbPort: process.env.POSTGRES_PORT || 5432,
  dbName: process.env.POSTGRES_DB || 'library',
  datadirs: {
    migrations: process.env.MIGRATIONS_DIR || './src/db/migrations',
    seed: process.env.SEED_DATA_DIR || './src/db/seed_data',
    corpora: process.env.CORPORA_DIR || './src/db/nltk_data/corpora/gutenberg'
  },
  clientAddr: process.env.CLIENT_ADDR || 'http://localhost:3000',
  rootUrl: process.env.ROOTURL || null
};
