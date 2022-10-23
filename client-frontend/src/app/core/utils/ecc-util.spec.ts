// import { EccUtil } from "./ecc-util"

describe('ECC Util', () => {
    it('populate pad this', () => {

    });
})
// describe('ECC Util', () => {
//     it('generates keys', () => {
//         EccUtil.generateKeys();
//     })
//     it('generates Ecdsa from base 64 private key', () => {
//         const jwk = {
//             kty: 'EC',
//             crv: 'P-256',
//             x: 'PCUqRPcJr7nkMEtTLgL9LURVJOnf7jMyY5DW09j5Ukc',
//             y: 'b15kClYehc4__j7gvXG5yWVRZqCSIujPAGXTbUa8toQ',
//             d: '_CvJJMidxZVC7J81eHk7REzK2y23qcssgikmx6t-tKs'
//         };
//         const result = EccUtil.generateEcdsa(jwk);
//         expect(result).toBeTruthy();
//     });
//     it('signs challenges', () => {
//         const jwk = {
//             kty: 'EC',
//             crv: 'P-256',
//             x: 'PCUqRPcJr7nkMEtTLgL9LURVJOnf7jMyY5DW09j5Ukc',
//             y: 'b15kClYehc4__j7gvXG5yWVRZqCSIujPAGXTbUa8toQ',
//             d: '_CvJJMidxZVC7J81eHk7REzK2y23qcssgikmx6t-tKs'
//         };
//         const result = EccUtil.signChallenge(jwk, 'hello_world!');
//         console.log(result);
//     })
// })