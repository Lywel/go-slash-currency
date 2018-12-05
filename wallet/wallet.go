package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	eth "github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"os"
)

func New(path string) (*ecdsa.PrivateKey, error) {
	// Check if wallet exists at specified path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// The wallet needs to be generated
		fmt.Println("Creating new wallet at " + path)
		key, err := ecdsa.GenerateKey(eth.S256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		err = savePEMKey(path, key)
		if err != nil {
			return nil, err
		}
		return key, nil
	}

	// a file exists at specified path
	return loadPEMKey(path)
}

func loadPEMKey(path string) (*ecdsa.PrivateKey, error) {
	fmt.Println("Loading wallet at " + path)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	PEMBlock, _ := pem.Decode(bytes)
	if PEMBlock == nil || PEMBlock.Type != "SLASH WALLET" {
		return nil, errors.New("wallet is of invalid type")
	}

	key, err := eth.ToECDSA(PEMBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func savePEMKey(fileName string, key *ecdsa.PrivateKey) error {
	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	//bytes, err := x509.MarshalECPrivateKey(key)
	bytes := eth.FromECDSA(key)
	if err != nil {
		return err
	}

	var privateKey = &pem.Block{
		Type:  "SLASH WALLET",
		Bytes: bytes,
	}

	err = pem.Encode(outFile, privateKey)
	if err != nil {
		return err
	}

	return nil
}
