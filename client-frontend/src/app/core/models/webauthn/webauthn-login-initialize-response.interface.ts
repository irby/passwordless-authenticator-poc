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