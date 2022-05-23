package registry

import (
	"github.com/docker/libtrust"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegistryToken_Generate(t *testing.T) {
	privateKey, errKey := libtrust.GenerateRSA2048PrivateKey()
	require.NoError(t, errKey)
	publicKey, errPubKey := libtrust.FromCryptoPublicKey(privateKey.CryptoPublicKey())
	require.NoError(t, errPubKey)

	rt := registryToken{
		tokenIssuer:     "OLYMP TESTER",
		tokenExpiration: 3,
		secret:          "super-test-secret",
		privateKey:      privateKey,
		publicKey:       publicKey,
	}
	authReq := AuthorizationRequest{
		Account: "Martian",
		Service: "127.0.0.1",
		Type:    "registry",
		Name:    "test-resource",
		Actions: []string{"pull", "push"},
	}

	jwtToken, err := rt.Generate(&authReq)
	assert.NoError(t, err)

	claims := jwt.MapClaims{}
	testToken, errToken := jwt.ParseWithClaims(jwtToken.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey.CryptoPublicKey(), nil
	})
	assert.NoError(t, errToken)
	assert.True(t, testToken.Valid)

	// generate with error

	_, err = rt.Generate(&authReq)
	assert.Error(t, err)
}
