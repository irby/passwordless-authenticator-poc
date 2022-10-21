import {ec} from 'elliptic';

export class EccService {
    public static signChallenge(privateKey: string, challenge: string): string {
        const EC = new ec('secp256k1');
        const key = EC.keyFromPrivate(privateKey, 'hex');

        const clientData = {
            type: "webauthn.get",
            challenge: this.sanitizeInput(""),
            origin: "http://localhost:4200"
        }

        const mashHash = Buffer.from(JSON.stringify(clientData), 'utf8');
        const signature = key.sign(mashHash);

        const hex = signature.r.toString('hex') + signature.s.toString('hex');
        const str = Buffer.from(hex, 'hex').toString('base64');

        return this.sanitizeInput(str);


        // return this.sanitizeInput(btoa(signature.toDER()));
    }

    private static sanitizeInput(challenge: string): string {
        return challenge.replace(/=/g, '').replace(/\//g, "_").replace(/\+/g, "-")
    }
}