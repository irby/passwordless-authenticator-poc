import { EccJwk } from "./services/ecc.service";

const mirby7Jwk: EccJwk = {
    kty: 'EC',
    crv: 'P-256',
    x: 'PCUqRPcJr7nkMEtTLgL9LURVJOnf7jMyY5DW09j5Ukc',
    y: 'b15kClYehc4__j7gvXG5yWVRZqCSIujPAGXTbUa8toQ',
    d: '_CvJJMidxZVC7J81eHk7REzK2y23qcssgikmx6t-tKs'
  };
const gburdell27Jwk: EccJwk =  {
    kty: 'EC',
    crv: 'P-256',
    x: 'fyXf7JFGCEnxSN0nQP1KkrWd2Dni9UBc6_b0PxQgf5o',
    y: 'TfzasgsU5K24pZ_2j-UGJ6QG38Z92vNtPsAVsdCRL9s',
    d: '3HW1lLIAu2A8TEmwfltV34cxHQP-4HLl-SlHtxN0YNw'
  }
const buzzJwk: EccJwk = {
    kty: 'EC',
    crv: 'P-256',
    x: 'KZPZEgrw195TmDSRzL3WQzTz7L2aTa1ArSLJTxV91-k',
    y: 'lJLVQJac1PrNzzKAP2U87iPrKdT9-GpjdHMuESjDKH8',
    d: 'O5IQnk3hzK0u7WtMZMy9ZysFLyDjzIpRmjCPS1GnlZU'
}

export function getJwkFromName(name: string): EccJwk {
    switch(name) {
        case "mirby7@gatech.edu":
            return mirby7Jwk;
        case "gburdell27@gatech.edu":
            return gburdell27Jwk;
        case "buzz@gatech.edu":
            return buzzJwk;
        default:
            throw new Error("user not found");
    }
}