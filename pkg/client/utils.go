package client

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

func ReadFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(file)
}

// func getPrivateKey(fileName string) (*rsa.PrivateKey, error) {

// 	pkey, err := ReadFile(fileName)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return getPrivateKeyFromBytes(pkey)
// }

func getPrivateKeyFromBytes(pkey []byte) (*rsa.PrivateKey, error) {
	kb, _ := pem.Decode(pkey)
	if kb == nil {
		return nil, fmt.Errorf("could not decode dkim key")
	}
	pk, err := x509.ParsePKCS1PrivateKey(kb.Bytes)
	if err != nil {
		return nil, err
	}
	return pk, nil
}
