import {ec} from 'elliptic';
import * as ECDSA from 'ecdsa-secp256r1';
import * as crypto from 'node:crypto';

export class EccService {
    public static signChallenge(privateKey: string, challenge: string): string {
        const EC = new ec('secp256k1');
        const key = EC.keyFromPrivate(privateKey, 'hex');

        console.log('challenge', challenge);

        const clientData = {
            type: "webauthn.get",
            challenge: this.sanitizeInput(challenge),
            origin: "http://localhost:4200"
        }

        console.log(clientData);

        const mashHash = Buffer.from(JSON.stringify(clientData), 'utf8');
        console.log(mashHash);
        const signature = key.sign(mashHash);

        console.log(signature.toDER());

        const hex = signature.r.toString('hex') + signature.s.toString('hex');
        const str = Buffer.from(hex, 'hex').toString('base64');

        return this.sanitizeInput(str);


        // return this.sanitizeInput(btoa(signature.toDER()));
    }

    public static signChallengeEcc(jwk: EccJwk, challenge: string): string {
        const ec = ECDSA.fromJWK(jwk);

        const authenticatorData = "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFAAAAAA";

        const clientData = {
            type: "webauthn.get",
            challenge: this.sanitizeInput(challenge),
            origin: "http://localhost:4200"
        };

        const buffer = Buffer.from(JSON.stringify(clientData));
        var byteArray = Buffer.from(authenticatorData, 'base64');
        console.log(byteArray);

        const clientDataHash = crypto.createHmac('sha256', buffer).digest('hex');
        const hashBytes = Buffer.from(clientDataHash, 'hex');
        const sigData = Array.prototype.concat(byteArray, hashBytes);

        console.log(sigData);

        const sigBuffer = Buffer.from(sigData);

        return ec.sign(sigBuffer);
    }

    private static sanitizeInput(challenge: string): string {
        return challenge.replace(/=/g, '').replace(/\//g, "_").replace(/\+/g, "-")
    }
}

export interface EccJwk {
    kty: string;
    crv: string;
    x: string;
    y: string;
    d: string;
}