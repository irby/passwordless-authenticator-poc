package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/stretchr/testify/assert"
)

func Test_GeneratePrivateKey(t *testing.T) {
	key, err := GeneratePrivateKey()
	assert.NoError(t, err)
	assert.True(t, key != nil)
}

func Test_GeneratePrivateKey_PrivateKeyToString(t *testing.T) {
	key, err := GeneratePrivateKey()
	assert.NoError(t, err)
	assert.True(t, key != nil)
}

func Test_GenerateEC2PublicKeyDataFromPrivateKey(t *testing.T) {
	key, err := GeneratePrivateKey()
	assert.NoError(t, err)
	assert.True(t, key != nil)
	str, err := GenerateEC2PublicKeyDataFromPrivateKey(*key)
	assert.NoError(t, err)

	// fmt.Printf("Private key: %s\nPublic key data: %s\nN: %s\n", key.D.String(), str, key.Curve.Params().N)

	result, err := base64.RawURLEncoding.DecodeString(str)
	assert.NoError(t, err)

	newKey, err := webauthncose.ParsePublicKey(result)
	assert.NoError(t, err)

	data := newKey.(webauthncose.EC2PublicKeyData)
	assert.Equal(t, int64(1), data.Curve)      // P256
	assert.Equal(t, int64(-7), data.Algorithm) // AlgES256
	assert.Equal(t, int64(2), data.KeyType)
	assert.Equal(t, 32, len(data.XCoord))
	assert.Equal(t, 32, len(data.YCoord))
}

func Test_GeneratePrivateKey_PrivateKeyFromValue(t *testing.T) {
	key, err := GeneratePrivateKey()
	assert.NoError(t, err)
	assert.True(t, key != nil)
	str := key.D.String()

	newKey, err := GeneratePrivateKeyFromValue(str)
	assert.NoError(t, err)
	assert.Equal(t, &newKey.D, &key.D)
}

func Test_SignData(t *testing.T) {
	key, err := GeneratePrivateKeyFromValue("16607140015661132309087522590752959541886570147214553558567331635599686272321")
	assert.NoError(t, err)
	_, err = SignData(key, []byte("hello"))
	assert.NoError(t, err)
	// fmt.Printf("%s\n", res)
}

func Test_GetClientData(t *testing.T) {
	challenge := "1234567890"
	data, err := GetClientData(challenge)
	assert.NoError(t, err)
	assert.True(t, data != nil)
	res := base64.URLEncoding.EncodeToString(data)
	assert.Equal(t, "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiMTIzNDU2Nzg5MCIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6NDIwMCJ9", res)
}

func Test_GetAuthenticatorData(t *testing.T) {
	res, _ := GetAuthenticatorData()
	assert.Equal(t, 37, len(res))
	array := []int{73, 150, 13, 229, 136, 14, 140, 104, 116, 52, 23, 15, 100, 118, 96, 91, 143, 228, 174, 185, 162, 134, 50, 199, 153, 92, 243, 186, 131, 29, 151, 99, 5, 0, 0, 0, 0}
	assert.Equal(t, len(array), len(res))
	for i := 0; i < len(res); i++ {
		assert.Equal(t, uint8(array[i]), res[i])
	}
}

func Test_SignChallengeForUser_Verification(t *testing.T) {
	key, err := GeneratePrivateKey()
	assert.NoError(t, err)

	challenge := "Z8jhcr6huZ03WKauWoz1xxsiZRDiWtT5Dy4OABMFT9k"
	signature, _ := SignChallengeForUser(key.D.String(), challenge)

	publicKeyStr, err := GenerateEC2PublicKeyDataFromPrivateKey(*key)
	assert.NoError(t, err)
	publicKey, err := base64.RawURLEncoding.DecodeString(publicKeyStr)
	assert.NoError(t, err)

	pubKey, err := webauthncose.ParsePublicKey(publicKey)
	assert.NoError(t, err)

	authenticatorData, _ := GetAuthenticatorData()
	clientData, _ := GetClientData(challenge)

	clientDataHash := sha256.Sum256(clientData)
	sigData := append(authenticatorData, clientDataHash[:]...)

	valid, err := webauthncose.VerifySignature(pubKey, sigData, signature)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func Test_E2E(t *testing.T) {
	keys, err := GeneratePrivateKey()
	assert.NoError(t, err)

	challenge := "Z8jhcr6huZ03WKauWoz1xxsiZRDiWtT5Dy4OABMFT9k"
	authenticatorData, _ := GetAuthenticatorData()
	clientData, err := GetClientData(challenge)
	assert.NoError(t, err)
	clientDataHash := sha256.Sum256(clientData)
	sigData := append(authenticatorData, clientDataHash[:]...)
	r, s, err := ecdsa.Sign(rand.Reader, keys, sigData)
	assert.NoError(t, err)
	isValid := ecdsa.Verify(&keys.PublicKey, sigData, r, s)
	assert.True(t, isValid)
}

func Test_E2E_3(t *testing.T) {
	keys, err := GeneratePrivateKey()
	assert.NoError(t, err)

	pubkey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     keys.X,
		Y:     keys.Y,
	}

	challenge := "Z8jhcr6huZ03WKauWoz1xxsiZRDiWtT5Dy4OABMFT9k"
	authenticatorData, _ := GetAuthenticatorData()
	clientData, err := GetClientData(challenge)
	assert.NoError(t, err)
	clientDataHash := sha256.Sum256(clientData)
	sigData := append(authenticatorData, clientDataHash[:]...)

	f := webauthncose.HasherFromCOSEAlg(webauthncose.COSEAlgorithmIdentifier(-7))
	h := f()
	h.Write(sigData)

	r, s, err := ecdsa.Sign(rand.Reader, keys, h.Sum(nil))
	assert.NoError(t, err)

	isValid := ecdsa.Verify(pubkey, h.Sum(nil), r, s)
	assert.True(t, isValid)
}
