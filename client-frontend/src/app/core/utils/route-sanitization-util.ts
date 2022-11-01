export class RouteSanitizationUtil {
    public static sanitizeRoute(route: string): RouteSanitizationResult {
        if (route?.length <= 0) {
            return {
                grantId: null,
                token: null
            };
        }
        const result: RouteSanitizationResult = {
            grantId: null,
            token: null
        };
        const temp = route?.split('?');
        result.grantId = temp[0];

        if (temp.length > 1) {
            let temp2 = temp[1].split('&');
            temp2.forEach(val => {
                if (!val.includes('token')) {
                    return;
                }
                result.token = val.replace('token=', '');
            });
        }

        return result
    }
}

export interface RouteSanitizationResult {
    grantId: string | null;
    token: string | null;
}