// file storage provider allow save and load RSA private key file to/from file
// It provider implemented keyStorage interface for using with key package
package storage

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
)

type FileStorage struct {
	rootPath   string // root path where files stored
	publicKey  string // a public key file
	privateKey string // a private key file
}

// NewFileStorage accept path to private and public key files
// File path doesn't check in constructor because files can be generate by 'key' package in future
func NewFileStorage(rootPath, privateKey, publicKey string) FileStorage {
	return FileStorage{
		rootPath:   rootPath,
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// Save will saving private and public Key to file
// A public key need save separately for using in isolated web-service as JWK,
// also save allow store certificates bundle files if they created
func (fs FileStorage) Save(key *rsa.PrivateKey, certCA []byte) error {
	privateKeyFile, err := os.Create(fmt.Sprintf("%s/%s", fs.rootPath, fs.privateKey))
	if err != nil {
		return errors.Wrap(err, "failed to save private key")
	}

	// close file on exit and check for it for error returned
	defer func() {
		if err := privateKeyFile.Close(); err != nil {
			log.Printf("failed to close private key file %v", err)
		}
	}()

	// processing with public key
	publicKeyFile, err := os.Create(fmt.Sprintf("%s/%s", fs.rootPath, fs.publicKey))
	if err != nil {
		return errors.Wrap(err, "failed to save private key")
	}

	defer func() {
		if err := publicKeyFile.Close(); err != nil {
			log.Printf("failed to close public key file %v", err)
		}
	}()
	b, err := getBytesPEM(key)
	if err != nil {
		return err
	}

	if _, err = privateKeyFile.Write(b); err != nil {
		return errors.Wrap(err, "failed to save private key to file")
	}

	b, err = getBytesPEM(key.Public())
	if err != nil {
		return err
	}

	if _, err = publicKeyFile.Write(b); err != nil {
		return errors.Wrap(err, "failed to save public key to file")
	}

	if certCA != nil {
		// processing with CA ROOT file
		caRootFile, err := os.Create(fmt.Sprintf("%s/CA_%s.%s", fs.rootPath, fs.publicKey, "crt"))
		if err != nil {
			return errors.Wrap(err, "failed to save private key")
		}

		defer func() {
			if err := caRootFile.Close(); err != nil {
				log.Printf("failed to close CA ROOT file %v", err)
			}
		}()
		if _, err = caRootFile.Write(certCA); err != nil {
			return errors.Wrap(err, "failed to save CA ROOT file")
		}
	}

	return nil
}

// Load will loading privateKey from PEM-file
func (fs FileStorage) Load() (*rsa.PrivateKey, error) {
	// path to private key file is required and throw error if path doesn't set
	if fs.privateKey == "" {
		return nil, errors.New("path to private key must be set")
	}

	privateKeyFile, err := os.Open(fmt.Sprintf("%s/%s", fs.rootPath, fs.privateKey))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load private key from file %s\\%s", fs.rootPath, fs.privateKey)
	}

	defer func() {
		if err = privateKeyFile.Close(); err != nil {
			log.Printf("failed to close private key file, err: %v", err)
		}
	}()

	privateKeyData, err := io.ReadAll(privateKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read data from private key file")
	}

	// before create private key need read PEM blocks from key data
	pemBlock, _ := pem.Decode(privateKeyData)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read pem block from private key data")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create private key from file data")
	}
	return privateKey.(*rsa.PrivateKey), nil
}

func getBytesPEM(key interface{}) ([]byte, error) {
	switch key.(type) {
	case *rsa.PrivateKey:
		keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, errors.New("failed to marshal private key to bytes")
		}
		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: keyBytes,
			},
		), nil
	case *rsa.PublicKey:
		keyBytes, err := x509.MarshalPKIXPublicKey(key)
		if err != nil {
			return nil, errors.Wrap(err, "failed parse public key to PEM bytes encode")
		}
		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: keyBytes,
			},
		), nil
	}
	return nil, errors.New("failed key file type for bytes encode")
}
