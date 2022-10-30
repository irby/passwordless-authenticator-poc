package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/stretchr/testify/assert"
)

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

	_ = base64.RawURLEncoding.EncodeToString(e.XCoord)
	_ = base64.RawURLEncoding.EncodeToString(e.YCoord)
	// fmt.Printf("x: %s\ny: %s\n", xString, yString)
}

func Test_BreakdownOfEccKey_Mirby7(t *testing.T) {
	xStr := "PCUqRPcJr7nkMEtTLgL9LURVJOnf7jMyY5DW09j5Ukc"
	yStr := "b15kClYehc4__j7gvXG5yWVRZqCSIujPAGXTbUa8toQ"

	x, err := base64.RawURLEncoding.DecodeString(xStr)
	assert.NoError(t, err)
	y, err := base64.RawURLEncoding.DecodeString(yStr)
	assert.NoError(t, err)

	e := webauthncose.EC2PublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			Algorithm: -7,
			KeyType:   2, // ECC
		},
		XCoord: x,
		YCoord: y,
		Curve:  1,
	}
	result, err := webauthncbor.Marshal(e)
	assert.NoError(t, err)
	key, err := webauthncose.ParsePublicKey(result)

	// fmt.Printf("%s\n", base64.RawURLEncoding.EncodeToString(result))

	data := key.(webauthncose.EC2PublicKeyData)
	assert.Equal(t, int64(1), data.Curve)      // P256
	assert.Equal(t, int64(-7), data.Algorithm) // AlgES256
	assert.Equal(t, int64(2), data.KeyType)
	assert.Equal(t, 32, len(data.XCoord))
	assert.Equal(t, 32, len(data.YCoord))
}

func Test_BreakdownOfSavedEccKey_Mirby7(t *testing.T) {
	public_key_string := "pQECAyYgASFYIHIEHgsag0GfyV8urYQE8fJoBRWhW_iBTp27kHVVcX9FIlggqodro9r_J1cynQ8PxyiGIcGF48B2V1FHSrUf7i9doAY"
	result, err := base64.RawURLEncoding.DecodeString(public_key_string)
	assert.NoError(t, err)
	key, err := webauthncose.ParsePublicKey(result)
	assert.NoError(t, err)
	data := key.(webauthncose.EC2PublicKeyData)
	assert.Equal(t, int64(1), data.Curve)      // P256
	assert.Equal(t, int64(-7), data.Algorithm) // AlgES256
	assert.Equal(t, int64(2), data.KeyType)
	assert.Equal(t, 32, len(data.XCoord))
	assert.Equal(t, 32, len(data.YCoord))
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

func stringToBin(s string) (binString string) {
	for _, c := range s {
		binString = fmt.Sprintf("%s%b", binString, c)
	}
	return
}
