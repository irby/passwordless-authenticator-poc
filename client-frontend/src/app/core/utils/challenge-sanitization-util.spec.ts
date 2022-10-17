import { ChallengeSanitizationUtil } from "./challenge-sanitization-util";

describe('ChallengeSanitizationUtil', () => {
    it('sanitizes all instance all of "=" to ""', () => {
        const input = "abc=defg=hijk==";
        const result = ChallengeSanitizationUtil.sanitizeInput(input);
        expect(result).toBe("abcdefghijk");
    })
})