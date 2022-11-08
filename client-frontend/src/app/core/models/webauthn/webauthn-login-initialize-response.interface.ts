export interface WebauthnLoginInitializeResponse {
    publicKey: PublicKey;
}

export interface PublicKey {
    challenge: string;
    timeout: number;
    rpId: string;
    allowedCredentials?: AllowedCredentials[];
    userVerification: string;
}

export interface AllowedCredentials {
    type: string;
    id: string;
}

export interface CreateAccountWithGrantResponse {
    options: WebauthnLoginInitializeResponse;
    grant: GrantAttestationObject;
}

export interface GrantAttestationObject {
    accountAccessGrantId: string;
    guestUserId: string;
    createdAt: Date;
    expireByTime: boolean;
    expireByLogins: boolean;
    minutesAllowed: number;
    loginsAllowed: number;
}