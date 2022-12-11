### JSON Web Key (JWK) tool
---
[![Build Status](https://github.com/zebox/gojwk/actions/workflows/main.yml/badge.svg)](https://github.com/zebox/gojwk/actions) [![Build Status](https://github.com/zebox/gojwk/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/zebox/gojwk/actions) [![Coverage Status](https://coveralls.io/repos/github/zebox/gojwk/badge.svg)](https://coveralls.io/github/zebox/gojwk) [![Go Report Card](https://goreportcard.com/badge/github.com/zebox/gojwk)](https://goreportcard.com/report/github.com/zebox/gojwk)

This simple library provides tools for work with private and public keys using RSA 
as [JWK](https://datatracker.ietf.org/doc/html/rfc7517).
The Library allows generating, save and load crypto keys pair based on RSA algorithm. 
JWKS usually use asymmetric encryption keys pair where public key (using in JWKS) for validate the 
[JWT](https://jwt.io/introduction) tokens which signed with private part of keys.
A public key can be placed at different service or server for validate JWT signature.

The Library write in Go and you can either embed to golang projects or use as a standalone application.

#### HOW TO USE
Main items of this library is crypto keys pair. You can generate they or load from some storage. Library supports both of this way (in currently support only RSA keys).

Init keys pair with `NewKeys`for create `Keys` instance.

Constructor can accept two options:
- Storage - this is interface which has `Load` and `Save` method. They define where keys will be stored and load from. 
User can use pre-defined storage `File` provider in `storage` package. By default, this option is undefined and new generated keys will store in memory only.
Storage `File` provider required path to private and public keys. 
  
- BitSize - defined size for crypto key which will be generated. Option accept `int` value  By default - 2048.

After `Keys` inited user should either `Generate` new key pair or `Load` from storage provider if keys doesn't exist in storage yet. 
  
```go
keys,err:=NewKeys() // if storage option undefined key pair store in memory
 
if err!=nil {
        // handle error 
}

// Generate new keys pair if need
if err=keys.Generate();err!=nil {
    // handle error
}

err,jwk:=keys.JWK()
if err!=nil {
    // handle error
}

fmt.Println(jwk.ToString())
```
A after execute code above you get result like this:
```javascript
{
          "kty": "RSA",
          "kid": "oI4f",
          "use": "sig",
          "alg": "RS256",
          "n": "n5Y24DhSDIKIN6tJbrOMxfZpoedvAIAA5vKv...",
          "e": "AQAB"
}
```
Example with options:
```go
// NewFileStorage accept rootPath, privateKey and publicKey names params
fs := storage.NewFileStorage("./","test_private.key", "test_public.key")
keys, err := NewKeys(Storage(fs))

// if storage provider hasn't keys pair yet user can generate they 
// after generated key pair will be save to defined storage
if err=keys.Generate();err!=nil {
    // handle error
}

// Load key pair from storage provider
if err=keys.Load();err!=nil {
    // handle error
}

err,jwk:=keys.JWK()
if err!=nil {
    // handle error
}
```
`Keys` has method `CreateCAROOT` for create Certificate Authority (CA) file with generated keys pair
```go
// create Keys instance
keys, err := NewKeys()
if err=keys.Load();err!=nil {
    // handle error
}

// create certificate data
ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{

			Organization:  []string{"TEST, INC."},
			Country:       []string{"RU"},
			Province:      []string{""},
			Locality:      []string{"Krasnodar"},
			StreetAddress: []string{"Krasnaya"},
			PostalCode:    []string{"350000"},
		},

		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// add Subject Alternative Name for requested IP and Domain
	// it prevent error with untrusted certificate for client request
	// https://oidref.com/2.5.29.17
	ca.IPAddresses = append(ca.IPAddresses, net.ParseIP("127.0.0.1"))
	ca.IPAddresses = append(ca.IPAddresses, net.ParseIP("::"))
	ca.DNSNames = append(ca.DNSNames, "localhost")

// generate RSA keys pair (private and public)
if err=keys.Generate();err!=nil {
    // handle error
}

// create CA certificate for created keys pair
if err = keys.CreateCAROOT(ca); err != nil {
	return nil, nil, err
}

// if storage provider defined user should call Save function for store certificate and keys files
```
Full example with web service usage see here [example](https://github.com/zebox/gojwk/blob/master/_example/main.go)

#### Status
The code still under development. Until v1.x released the API & protocol may change.



 