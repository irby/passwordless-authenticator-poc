package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

//func TestPublicKey_Parse(t *testing.T) {
//	var strVal = "pQED"
//	var strValEnd = "ASFYIHIEHgsag0GfyV8urYQE8fJoBRWhW_iBTp27kHVVcX9FIlggqodro9r_J1cynQ8PxyiGIcGF48B2V1FHSrUf7i9doAY"
//	//var strVal = ""
//	//var strValEnd = ""
//	fmt.Println("Here")
//	i := 0
//	for i < 1 {
//		temp := shortID(103 - len(strValEnd) - len(strVal))
//		pk := webauthncose.PublicKeyData{}
//		byteCredentialPubKey, _ := base64.RawURLEncoding.DecodeString(strVal + temp + strValEnd)
//		err := webauthncbor.Unmarshal(byteCredentialPubKey, &pk)
//		if err != nil {
//			continue
//		}
//		if webauthncose.COSEKeyType(pk.KeyType) != webauthncose.RSAKey {
//			continue
//		}
//		res := webauthncose.COSEAlgorithmIdentifier(pk.Algorithm)
//		if pk.Algorithm == 0 {
//			continue
//		}
//		if res == webauthncose.AlgES512 || res == webauthncose.AlgES384 || res == webauthncose.AlgES256 {
//			continue
//		}
//		if res == webauthncose.AlgRS1 || res == webauthncose.AlgPS256 || res == webauthncose.AlgRS256 || res == webauthncose.AlgPS512 || res == webauthncose.AlgRS512 {
//			fmt.Println(pk.Algorithm, webauthncose.COSEAlgorithmIdentifier(pk.Algorithm))
//			strVal = strVal + temp + strValEnd
//			i = 1
//			break
//		}
//	}
//	fmt.Println(strVal)
//	//pk := PublicKeyData{}
//	//byteCredentialPubKey, _ := base64.RawURLEncoding.DecodeString("pQMmIAEhWCAoCF-x0dwEhzQo-ABxHIAgr_5WL6cJceREc81oIwFn7iJYIHEHx8ZhBIE42L26-rSC_3l0ZaWEmsHAKyP9rgslApUdAQI")
//	//err := webauthncbor.Unmarshal(byteCredentialPubKey, &pk)
//	//if err != nil {
//	//	// Handle
//	//}
//	//if webauthncose.COSEKeyType(pk.KeyType) != webauthncose.RSAKey {
//	//	// Handle
//	//}
//	//if webauthncose.COSEAlgorithmIdentifier(pk.Algorithm) != webauthncose.AlgRS256 {
//	//	// Handle
//	//}
//}

func TestBreakdownOfEccKey(t *testing.T) {
	public_key := "pQECAyYgASFYIHIEHgsag0GfyV8urYQE8fJoBRWhW_iBTp27kHVVcX9FIlggqodro9r_J1cynQ8PxyiGIcGF48B2V1FHSrUf7i9doAY"
	public_key_bytes, err := base64.RawURLEncoding.DecodeString(public_key)
	assert.NoError(t, err)
	_, err = webauthncose.ParsePublicKey(public_key_bytes)
	assert.NoError(t, err)

	pk := webauthncose.PublicKeyData{}

	err = webauthncbor.Unmarshal(public_key_bytes, &pk)
	assert.NoError(t, err)

	conv := webauthncose.COSEKeyType(pk.KeyType)
	assert.Equal(t, webauthncose.EllipticKey, conv)

	val := webauthncose.COSEAlgorithmIdentifier(pk.Algorithm)
	assert.Equal(t, webauthncose.AlgES256, val)

	var e webauthncose.EC2PublicKeyData

	err = webauthncbor.Unmarshal(public_key_bytes, &e)
	assert.NoError(t, err)

	xString := base64.RawURLEncoding.EncodeToString(e.XCoord)
	assert.NoError(t, err)
	yString := base64.RawURLEncoding.EncodeToString(e.YCoord)
	assert.NoError(t, err)
	fmt.Printf("x: %s\ny: %s\n", xString, yString)
}

func Test_BreakdownOfEccKey_Mirby7(t *testing.T) {
	xStr := "d6b5d57af9afef9f3fb31ccf532d25338570c86f48209b43d8119a55358c8b5c"
	yStr := "27f0b0210f21c9f6558adc6ef07653bc04a3bcc1c3a4fadf429c2f075c499229"

	xInt, isAssigned := new(big.Int).SetString(xStr, 16)
	assert.True(t, isAssigned)

	yInt, isAssigned := new(big.Int).SetString(yStr, 16)
	assert.True(t, isAssigned)

	e := webauthncose.EC2PublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			Algorithm: -7,
			KeyType:   2, // ECC
		},
		XCoord: xInt.Bytes(),
		YCoord: yInt.Bytes(),
		Curve:  1,
	}
	result, err := webauthncbor.Marshal(e)
	assert.NoError(t, err)
	key, err := webauthncose.ParsePublicKey(result)

	fmt.Printf("%s\n", base64.RawURLEncoding.EncodeToString(result))
	assert.Equal(t, int64(1), key.(webauthncose.EC2PublicKeyData).Curve)
}

//func Test_BreakdownOfRsaKey_Mirby7(t *testing.T) {
//	modulus := "dc7c0baeb24174f2bd8129c046627883d888519fbda8244afed829341a73763c6e5d6095e8bd90c861edc19b0b5d844e7cc40defb0c0235ccfd4a545665b0ba359db4118c010cb9fc74312b213f237eac6ba4491997610be7b51364b2a8fe304adf1a8d3623aff097da41acf7629194f6c0e054b0e9c4a6c3ef32a1354d5a61b"
//	modInt, isAssigned := new(big.Int).SetString(modulus, 16)
//	assert.True(t, isAssigned)
//
//	exponent := "10001"
//	exponentInt, isAssigned := new(big.Int).SetString(exponent, 16)
//	assert.True(t, isAssigned)
//	r := webauthncose.RSAPublicKeyData{
//		PublicKeyData: webauthncose.PublicKeyData{
//			Algorithm: -257, //AlgRS256
//			KeyType:   3,    //RSA
//		},
//		Modulus:  modInt.Bytes(),
//		Exponent: exponentInt.Bytes(),
//	}
//	result, err := webauthncbor.Marshal(r)
//	assert.NoError(t, err)
//	_, err = webauthncose.ParsePublicKey(result)
//	assert.NoError(t, err)
//	fmt.Printf("%s\n", base64.RawURLEncoding.EncodeToString(result))
//
//	_ = "zl30Fgl_GDqK8AP3Z9E_Mm8X3_SwRKipMOKNQelbBVQ" // challenge
//
//	//_, err = base64.URLEncoding.DecodeString("QgWelEY1HVd8z3_6lo2hK4slFgq_ZIiVBarotVN95-H6kJO9LEC6CurNaCyJujSulFzgop9XOi1QYFMwLiLf8jlEJGOykt6BA92VKd0ro8HOF1VJAYAYxBQ99aTVp_Z1keDrVoWfs8brODvtTkdwi1p7DM6At8P6SMlgeH0fq-4")
//	//assert.NoError(t, err)
//	//clientDataJson, err := base64.URLEncoding.DecodeString("eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiemwzMEZnbF9HRHFLOEFQM1o5RV9NbThYM19Td1JLaXBNT0tOUWVsYkJWUSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6NDIwMCJ9")
//	//assert.NoError(t, err)
//	//fmt.Println(clientDataJson)
//	//authData, err := base64.URLEncoding.DecodeString("SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFAAAAAA")
//	//assert.NoError(t, err)
//	//fmt.Println(authData)
//	//clientDataHash := sha256.Sum256(clientDataJson)
//	//sigData := append(authData, clientDataHash[:]...)
//	//valid, err := webauthncose.VerifySignature(key, sigData, signature)
//	//assert.NoError(t, err)
//	//assert.True(t, valid)
//}
//
//func Test_BreakdownOfRsaKey_Gburdell27(t *testing.T) {
//	r := webauthncose.RSAPublicKeyData{
//		PublicKeyData: webauthncose.PublicKeyData{
//			Algorithm: -257, //AlgRS256
//			KeyType:   3,    //RSA
//		},
//		Modulus:  []byte("168031827384520219244810800583526129270770826541604041494525561150688078853196075301058892991414246335851143559089650736155557663901961219357740100290673124767384538995263025525714622371142112094288420920293008569787271558102368709066320446627157457706764317402072733295654572081189278699370860311136853206193"),
//		Exponent: []byte("65537"),
//	}
//	result, err := webauthncbor.Marshal(r)
//	assert.NoError(t, err)
//	_, err = webauthncose.ParsePublicKey(result)
//	assert.NoError(t, err)
//	fmt.Printf("%s\n", base64.RawURLEncoding.EncodeToString(result))
//}
//
//func Test_BreakdownOfRsaKey_Buzz(t *testing.T) {
//
//	r := webauthncose.RSAPublicKeyData{
//		PublicKeyData: webauthncose.PublicKeyData{
//			Algorithm: -257, //AlgRS256
//			KeyType:   3,    //RSA
//		},
//		Modulus:  []byte("151577161380419228029867984733607288373912064283010388497954259940298334865359507170487311958365840679182385033245409830274596853878402175872868162131433562795365230535759574619536454952510333900868091781977034498007456591937359485634578174398312391680700977718698186645288854739540537150863744303697546795753"),
//		Exponent: []byte("65537"),
//	}
//	result, err := webauthncbor.Marshal(r)
//	assert.NoError(t, err)
//	_, err = webauthncose.ParsePublicKey(result)
//	assert.NoError(t, err)
//	fmt.Printf("%s\n", base64.RawURLEncoding.EncodeToString(result))
//}

func shortID(length int) string {
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_"
	ll := len(chars)
	b := make([]byte, length)
	rand.Read(b) // generates len(b) random bytes
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%ll]
	}
	return string(b)
}

func stringToBin(s string) (binString string) {
	for _, c := range s {
		binString = fmt.Sprintf("%s%b", binString, c)
	}
	return
}
