import express from 'express';
import ec from 'elliptic';
const app = express();
const port = 3000;

app.get('/', (req, res) => {
    var EC = new ec.ec('secp256k1');
    var keyPair = EC.genKeyPair();

    var privateKey = keyPair.getPrivate();
    var publicKey = keyPair.getPublic();
    var publicKeyEnc = keyPair.getPublic('hex');
    var x = publicKey.getX();
    var y = publicKey.getY();
    console.log(x);
    console.log(y);
    console.log(privateKey);
    console.log('priv', privateKey.toString('hex'))
    console.log('x', x.toString('hex'));
    console.log('y', y.toString('hex'));
    console.log('curve', EC.curve.type);
    var msgHash = [ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 ];
    const hash1 = keyPair.sign(msgHash);
    console.log(hash1.toDER());


    const key = EC.keyFromPrivate(privateKey.toString('hex'), 'hex');
    const hash2 = keyPair.sign(msgHash);
    console.log(hash2.toDER(), hash1.toDER().toString('hex') === hash2.toDER().toString('hex'));
    console.log(btoa(hash2.toDER()));
  res.send('Hello World!!');
});

app.listen(port, () => {
  return console.log(`Express is listening at http://localhost:${port}`);
});
