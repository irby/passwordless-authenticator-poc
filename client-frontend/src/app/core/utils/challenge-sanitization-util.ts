export class ChallengeSanitizationUtil {
    static sanitizeInput(challenge: string): string {
        return challenge.replace(/=/g, '').replace(/\//g, "_").replace(/\+/g, "-");
    }
}