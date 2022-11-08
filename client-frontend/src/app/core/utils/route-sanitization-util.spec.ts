import { RouteSanitizationUtil } from "./route-sanitization-util";

describe("Route Sanitization Util", () => {
  it("returns share ID and token if both in route", () => {
    const route =
      "e9ef080-635b-4733-bee7-8bdae4757810?token=thisisatoken-listentome";
    const result = RouteSanitizationUtil.sanitizeRoute(route);
    expect(result.grantId).toBe("e9ef080-635b-4733-bee7-8bdae4757810");
    expect(result.token).toBe("thisisatoken-listentome");
  });
  it("returns share ID and token if other param in route", () => {
    const route =
      "e9ef080-635b-4733-bee7-8bdae4757810?token=thisisatoken-listentome&otherval=something-else";
    const result = RouteSanitizationUtil.sanitizeRoute(route);
    expect(result.grantId).toBe("e9ef080-635b-4733-bee7-8bdae4757810");
    expect(result.token).toBe("thisisatoken-listentome");
  });
  it("returns share ID and token if params re-arranged", () => {
    const route =
      "e9ef080-635b-4733-bee7-8bdae4757810?otherval=something-else&token=thisisatoken-listentome";
    const result = RouteSanitizationUtil.sanitizeRoute(route);
    expect(result.grantId).toBe("e9ef080-635b-4733-bee7-8bdae4757810");
    expect(result.token).toBe("thisisatoken-listentome");
  });
  it("returns share ID and null if only share ID in route", () => {
    const route = "e9ef080-635b-4733-bee7-8bdae4757810";
    const result = RouteSanitizationUtil.sanitizeRoute(route);
    expect(result.grantId).toBe("e9ef080-635b-4733-bee7-8bdae4757810");
    expect(result.token).toBe(null);
  });
  it("returns share ID and null if share ID and unknown param in route", () => {
    const route = "e9ef080-635b-4733-bee7-8bdae4757810?hello=woroldlowlwlod";
    const result = RouteSanitizationUtil.sanitizeRoute(route);
    expect(result.grantId).toBe("e9ef080-635b-4733-bee7-8bdae4757810");
    expect(result.token).toBe(null);
  });
  it("returns null null if string is blank", () => {
    const route = "";
    const result = RouteSanitizationUtil.sanitizeRoute(route);
    expect(result.grantId).toBe(null);
    expect(result.token).toBe(null);
  });
});
