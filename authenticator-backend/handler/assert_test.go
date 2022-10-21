package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/stretchr/testify/assert"
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
