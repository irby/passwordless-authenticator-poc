import express, { json, Router } from 'express';
import bodyParser from 'body-parser';
import { EccService } from './services/ecc.service';
import {getPrivateKeyByName} from './keys';
import {ec} from 'elliptic';
import {ResolveChallengeRequest} from './models/resolve-challenge-request';
const app = express();
const port = 3000;

// create application/json parser
var jsonParser = bodyParser.json()
 
// create application/x-www-form-urlencoded parser
var urlencodedParser = bodyParser.urlencoded({ extended: false })

// app.get('/', (req, res) => {
//     var EC = new ec('secp256k1');
//     var keyPair = EC.genKeyPair();

//     var privateKey = keyPair.getPrivate();
//     var publicKey = keyPair.getPublic();
//     var x = publicKey.getX();
//     var y = publicKey.getY();
//     // console.log(x);
//     // console.log(y);
//     // console.log(privateKey);
//     // console.log('priv', privateKey.toString('hex'))
//     // console.log('x', x.toString('hex'));
//     // console.log('y', y.toString('hex'));
//     // console.log('curve', EC.curve.type);

//     const clientData = {
//         type: "webauthn.get",
//         challenge: sanitizeInput("0SYebf4JjfD+nN589Lt4nB0fIa08rwdRQH8+agqpGJo="),
//         origin: "http://localhost:4200"
//     };



//     var msgHash = Buffer.from(JSON.stringify(clientData), 'utf8');
//     const hash1 = keyPair.sign(msgHash);
//     const hex = (hash1.r || hash1.s).toString('hex')
//     console.log(Buffer.from(hex, 'hex').toString('base64'))
//     console.log(hash1.r.toString('hex')+hash1.s.toString('hex'));

//     function sanitizeInput(challenge: string): string {
//         return challenge.replace(/=/g, '').replace(/\//g, "_").replace(/\+/g, "-")
//     }
//   res.send('Hello World!!');
// });

app.use((req, res, next) => {
    const allowedOrigins = ["http://localhost:4200"];
    const origin = req.headers.origin;

    res.setHeader('Access-Control-Allow-Origin', "*");

    res.setHeader('Access-Control-Allow-Credentials', 'true');
    res.setHeader('Access-Control-Allow-Headers', 'Host, Referer, User-Agent, Origin, Access-Control, Allow-Origin, Content-Type, Accept, Authorization, Origin, Accept-Encoding, Accept-Language, X-Requested-With, Access-Control-Request-Method, Access-Control-Request-Header, User-Id');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS, PUT, DELETE');

    next();
})

app.get('/', (req, res) => {
    res.send('Hello World!');
});

app.post('/', jsonParser, (req, res) => {
    const request = <ResolveChallengeRequest> req.body;

    if (request.email?.trim() === "" || request.challenge?.trim() === "") {
        return res.status(400).json("email and challenge are required");
    }

    const key = getPrivateKeyByName(request.email);
    const signature = EccService.signChallenge(key, request.challenge);
    res.send({"signature": signature});
});


app.listen(port, () => {
  return console.log(`Express is listening at http://localhost:${port}`);
});
