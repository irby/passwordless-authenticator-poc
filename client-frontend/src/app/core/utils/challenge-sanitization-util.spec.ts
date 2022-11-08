import { ChallengeSanitizationUtil } from "./challenge-sanitization-util";

describe("ChallengeSanitizationUtil", () => {
  it('sanitizes all instance all of "=" to ""', () => {
    const input = "abc=defg=hijk==";
    const result = ChallengeSanitizationUtil.sanitizeInput(input);
    expect(result).toBe("abcdefghijk");
  });

  it('sanitizes all instance all of "/" to "_"', () => {
    const input = "abc/defg/hijk/";
    const result = ChallengeSanitizationUtil.sanitizeInput(input);
    expect(result).toBe("abc_defg_hijk_");
  });

  it('sanitizes all instance all of "+" to "-"', () => {
    const input = "abc+defg+hijk++";
    const result = ChallengeSanitizationUtil.sanitizeInput(input);
    expect(result).toBe("abc-defg-hijk--");
  });

  it("sanitizes all instances of special characters to their converted values", () => {
    const input = "abc==defg/h/ijk++";
    const result = ChallengeSanitizationUtil.sanitizeInput(input);
    expect(result).toBe("abcdefg_h_ijk--");
  });
});
