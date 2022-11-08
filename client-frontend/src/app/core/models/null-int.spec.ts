import { NullInt, NullInt32 } from "./null-int";

describe("NullInt", () => {
  it("should return integer from getValueOrDefault when valid is true", () => {
    const conv = new NullInt(10, true);
    const result = conv.getValueOrDefault();
    expect(result).toBe(10);
  });
  it("should return null from getValueOrDefault when valid is false", () => {
    const conv = new NullInt(10, false);
    const result = conv.getValueOrDefault();
    expect(result).toBe(null);
  });
});
