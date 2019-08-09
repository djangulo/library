const cors = require("cors");

const express = require("express");
const path = require("path");

const config = require("./config");
const routes = require("./routes");

const { servePort, rootUrl } = config;

const appRoutes = [
  { path: "/books", router: routes.books },
  { path: "/pages", router: routes.pages }
];

const { dbName, dbUser, dbHost, dbPass, dbPort, datadirs } = config;

const app = express();
app.use(cors());
app.use("/books", routes.books);
app.use("/pages", routes.pages);

app.get("/", (req, res) => {
  res.sendFile(path.join(__dirname + "/index.html"));
});

app.listen(servePort, () =>
  console.log(
    "Library API listening on " + (rootUrl ? rootUrl : "port :" + servePort)
  )
);
