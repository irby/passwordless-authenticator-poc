import { GetPrivateKeyFromId, GetPublicKeyFromId, UserId } from "../constants/user-constants";
import { CryptoUtil } from "./crypto-util";

describe('CryptoUtil', () => {
    it('sign works', () => {
        const privateKey = GetPrivateKeyFromId(UserId.mirby7Id) as string;
        const dataToSign = {
            text: "hello world!"
        };
        const signature = CryptoUtil.signObject(privateKey, dataToSign);
        console.log(Buffer.from(signature).toString('base64'));
        expect(signature.length).toBeGreaterThan(0);
    });

    it('verify works', () => {
        const privateKey = GetPrivateKeyFromId(UserId.mirby7Id) as string;
        const publicKey = GetPublicKeyFromId(UserId.mirby7Id) as string;
        const dataToSign = {
            text: "hello world!"
        };
        const signature = CryptoUtil.signObject(privateKey, dataToSign);
        const verify = CryptoUtil.verifyObject(publicKey, dataToSign, signature);
        expect(verify).toBeTruthy();
    });
});
