import isEqual from "lodash/isEqual";
import { paginate } from "./paginate";

describe("paginate fn tests", () => {
  const items = Array.from(Array(30).keys());
  const cases = [
    {
      name: "takes the proper amount for the first page",
      in: { items, number: 1, size: 10 },
      want: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    },
    {
      name: "handles empty array",
      in: { items: [], number: 1, size: 10 },
      want: []
    },
    {
      name: "handles n=1 array",
      in: { items: [1], number: 1, size: 10 },
      want: [1]
    },
    {
      name: "handles page bigger than available by returning last page",
      in: { items, number: 4, size: 10 },
      want: [20, 21, 22, 23, 24, 25, 26, 27, 28, 29]
    },
    {
      name: "handles remainder last page",
      in: { items, number: 5, size: 7 },
      want: [28, 29]
    }
  ];
  for (let c of cases) {
    it(c.name, () => {
      let got = paginate(c.in.items, c.in.number, c.in.size);
      if (c.name === "handles remainder last page") {
        if (!isEqual(got, c.want)) console.log(`got ${got} want ${c.want}`);
      }
      expect(isEqual(got, c.want)).toBe(true);
    });
  }
});
