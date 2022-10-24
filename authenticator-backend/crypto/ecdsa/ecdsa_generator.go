package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"math/big"
)

const (
	Mirby7PrivateKey     = "114171996321272141621774891909004856173651256657292996959166070546390288808945"
	Mirby7PublicKey      = "pQECAyYgASFYIAQ1U2ym91-HiRzc4Og8mxTpmtJTMvErSvL8PMXrWugNIlggGGtZ38x0-4_nlxkkAQ-dzaGtia-5KcyxTBduawLxDF8"
	Gburdell27PrivateKey = "77224901188531167258934258406051002543979566815437818769772152906821540243761"
	Gburdell27PublicKey  = "pQECAyYgASFYIFsogqBB9bkmgJWZ5UmO_alfoLmtVDihMQGevHst9v0FIlggA6S4ercOjmzyhN-glwStlrvcdLq8NQ29FwLy75CTock"
	BuzzPrivateKey       = "89363970210854662954731113938582766963715078670905376571054837474978342770273"
	BuzzPublicKey        = "pQECAyYgASFYICdkIyxOAovuqIwI-ke2lzOVcvHIMs9NWURskKTvs98mIlgg4_vTn0-ZAmKiBtok22LFpD7SKX9zVVNSXbE397eMy6s"
)

type GeneratePrivateAndPublicKeyResponse struct {
	privateKey ecdsa.PrivateKey
}

func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func GenerateEC2PublicKeyDataFromPrivateKey(key ecdsa.PrivateKey) (string, error) {
	x := key.X
	y := key.Y

	e := webauthncose.EC2PublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			Algorithm: -7, // AlgES256
			KeyType:   2,  // ECC
		},
		XCoord: x.Bytes(),
		YCoord: y.Bytes(),
		Curve:  1,
	}

	result, err := webauthncbor.Marshal(e)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(result), nil
}

func GeneratePrivateKeyFromValue(privateKey string) (*ecdsa.PrivateKey, error) {
	newInt, isAssigned := new(big.Int).SetString(privateKey, 10)
	if !isAssigned {
		return nil, errors.New("unable to parse bigint from private key")
	}

	newKey := ecdsa.PrivateKey{
		D: newInt,
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
	}
	return &newKey, nil
}

type ECDSASignature struct {
	R, S *big.Int
}

func SignData(key *ecdsa.PrivateKey, data []byte) ([]byte, error) {

	f := webauthncose.HasherFromCOSEAlg(webauthncose.COSEAlgorithmIdentifier(-7))
	h := f()
	h.Write(data)

	res, err := ecdsa.SignASN1(rand.Reader, key, h.Sum(nil))
	if err != nil {
		return nil, err
	}

	return res, nil
}

type ClientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

func GetClientData(challenge string) ([]byte, error) {
	data := ClientData{
		Type:      "webauthn.get",
		Challenge: challenge,
		Origin:    "http://localhost:4200",
	}
	res, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetAuthenticatorData() ([]byte, error) {
	data := "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFAAAAAA"
	result, err := base64.URLEncoding.DecodeString(data)
	result = append(result, byte(0x00))
	return result, err
}

func SignChallengeForUser(privateKeyString string, challenge string) ([]byte, error) {
	key, err := GeneratePrivateKeyFromValue(privateKeyString)
	if err != nil {
		return nil, err
	}
	authenticatorData, _ := GetAuthenticatorData()
	clientData, err := GetClientData(challenge)
	if err != nil {
		return nil, err
	}
	clientDataHash := sha256.Sum256(clientData)
	sigData := append(authenticatorData, clientDataHash[:]...)
	data, err := SignData(key, sigData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//func getPrivateKeyForUser(user string) (string, error) {
//	var privateKey string
//	switch user {
//	case "mirby7@gatech.edu":
//		privateKey = Mirby7PrivateKey
//	case "gburdell27@gatech.edu":
//		privateKey = Gburdell27PrivateKey
//	case "buzz@gatech.edu":
//		privateKey = BuzzPrivateKey
//	default:
//		return "", fmt.Errorf("unable to find user %s", user)
//	}
//	return privateKey, nil
//}

func getPublicKeyForUser(user string) (string, error) {
	var privateKey string
	switch user {
	case "mirby7@gatech.edu":
		privateKey = Mirby7PublicKey
	case "gburdell27@gatech.edu":
		privateKey = Gburdell27PublicKey
	case "buzz@gatech.edu":
		privateKey = BuzzPublicKey
	default:
		return "", fmt.Errorf("unable to find user %s", user)
	}
	return privateKey, nil
}
