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

        console.log(JSON.stringify(clientData));

        const authenticatorData = "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFAAAAAA";

        const buffer = Buffer.from(JSON.stringify(clientData));

        console.log('client data JSON buffer', buffer);

        var byteArray = Buffer.from(authenticatorData, 'base64');
        console.log(byteArray);

        const clientDataHash = crypto.createHash('sha256').update(buffer).digest('hex');
        // const clientDataHash = crypto.createHmac('sha256', buffer).digest('hex');

        console.log('client data hash', clientDataHash);

        const hashBytes = Buffer.from(clientDataHash, 'hex');
        Buffer.concat([byteArray, hashBytes])
        const sigData = Buffer.concat([byteArray, hashBytes]);

        // const sigBuffer = Buffer.from(sigData);

        console.log('buff', sigData);

        for (var i = 0; i < sigData.length; i++) {
            console.log(sigData[i]);
        }

        console.log(clientData);

        const signature = key.sign(sigData);

        console.log(signature);

        console.log(signature.toDER());

        const hex = signature.r.toString('hex') + signature.s.toString('hex');
        const str = Buffer.from(hex, 'hex').toString('base64');

        return this.sanitizeInput(str);


        // return this.sanitizeInput(btoa(signature.toDER()));
    }

    public static signChallengeEcc(jwk: EccJwk, challenge: string): string {
        const ec = ECDSA.fromJWK(jwk);

        console.log("challenge", challenge)

        const authenticatorData = "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFAAAAAA";

        const clientData = {
            type: "webauthn.get",
            challenge: this.sanitizeInput(challenge),
            origin: "http://localhost:4200"
        };

        console.log(JSON.stringify(clientData));

        const buffer = Buffer.from(JSON.stringify(clientData));

        console.log('client data JSON buffer', buffer);

        var byteArray = Buffer.from(authenticatorData, 'base64');
        console.log(byteArray);

        const clientDataHash = crypto.createHash('sha256').update(buffer).digest('hex');
        // const clientDataHash = crypto.createHmac('sha256', buffer).digest('hex');

        console.log('client data hash', clientDataHash);

        const hashBytes = Buffer.from(clientDataHash, 'hex');
        Buffer.concat([byteArray, hashBytes])
        const sigData = Buffer.concat([byteArray, hashBytes]);

        // const sigBuffer = Buffer.from(sigData);

        console.log('buff', sigData);

        for (var i = 0; i < sigData.length; i++) {
            console.log(sigData[i]);
        }

        const sign = crypto.createSign('RSA-SHA256');
        sign.write(sigData);
        sign.end();
        const d = sign.sign(ec.toPEM());

        console.log(d);

        const signature = ec.sign(sigData);

        console.log('signature', signature);

        return signature;
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