require("@babel/polyfill");
const request = require("supertest");

const createServer = require("./index");

describe("GET /", () => {
  test("it should redirect to config.clientAddr", async () => {
    const server = await createServer();
    const response = await request(server).get("/");
    expect(response.statusCode).toBe(301);
  });
});
