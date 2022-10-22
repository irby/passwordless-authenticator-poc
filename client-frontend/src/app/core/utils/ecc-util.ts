// import * as ECDSA from 'ecdsa-secp256r1';
// export class EccUtil {
//     public static generateKeys() {
//         const key = ECDSA.generateKey();
//         console.log(key.asPublic().toJWK());
//         console.log(key.toJWK());
//         console.log(key.toPEM());
//     }

//     public static generateEcdsa(jwk: EccJwk) {
//         return ECDSA.fromJWK(jwk);
//     }

//     public static signChallenge(jwk: EccJwk, challenge: string): string {
//         const ec = this.generateEcdsa(jwk);
//         return ec.sign(challenge);
//     }
// }

// export interface EccJwk {
//     kty: string;
//     crv: string;
//     x: string;
//     y: string;
//     d: string;
// }