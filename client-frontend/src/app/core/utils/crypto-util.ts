import * as node from 'node-forge';
export class CryptoUtil {
    static async generateKeyPair(): Promise<void> {
    //    generateKeyPair('rsa', {
    //         modulusLength: 1024,
    //         publicKeyEncoding: {
    //         type: 'spki',
    //         format: 'pem'
    //         },
    //         privateKeyEncoding: {
    //         type: 'pkcs8',
    //         format: 'pem'
    //         }
    //     }, (err, publicKey, privateKey) => {
    //         console.log(publicKey);
    //         console.log(privateKey);
    //     });
    }

    static signObject(privateKey: string, data: any): string {
        const digest = node.md.sha256.create();
        digest.update(JSON.stringify(data), 'utf8');
        const key = node.pki.privateKeyFromPem(privateKey);
        return key.sign(digest);
        // const signerObject = createSign("RSA-SHA256");
        // signerObject.update(JSON.stringify(data));
        // var signature = signerObject.sign({key: privateKey, padding: constants.RSA_PKCS1_PSS_PADDING}, "base64");
        // return signature;
    }

    static verifyObject(publicKey: string, data: any, signature: string): boolean {
        const key = node.pki.publicKeyFromPem(publicKey);
        const digest = node.md.sha256.create();
        digest.update(JSON.stringify(data), 'utf8');
        return key.verify(digest.digest().bytes(), signature);
        // const verifierObject = createVerify("RSA-SHA256");
        // verifierObject.update(JSON.stringify(data));
        // return verifierObject.verify({key: publicKey, padding: constants.RSA_PKCS1_PSS_PADDING}, signature, "base64");
    }
}