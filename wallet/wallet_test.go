package wallet_test

import (
	"bitbucket.org/ventureslash/go-slash-currency/wallet"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func sign(data []byte, privkey *ecdsa.PrivateKey) ([]byte, error) {
	hashData := crypto.Keccak256([]byte(data))
	return crypto.Sign(hashData, privkey)
}

func verif(data []byte, pubKey *ecdsa.PublicKey, sig []byte) bool {
	hashData := crypto.Keccak256([]byte(data))
	recoveredPubkey, err := crypto.SigToPub(hashData, sig)
	if err != nil || recoveredPubkey == nil {
		panic("Signature verification failed: " + err.Error())
	}

	return reflect.DeepEqual(pubKey, recoveredPubkey)
}

func TestSaveAndLoadWallet(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-*.wallet")
	if err != nil {
		panic(err)
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name())

	w1, err := wallet.New(tmpfile.Name())
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	w2, err := wallet.New(tmpfile.Name())
	if err != nil {
		panic(err)
	}

	// w1 should be the same as w2
	data := []byte("testing data")

	sig1, err := sign(data, w1)
	if err != nil {
		panic(err)
	}

	if !verif(data, &w2.PublicKey, sig1) {
		t.Fatal("Sig1 can't be checked with w2")
	}

	sig2, err := sign(data, w2)
	if err != nil {
		panic(err)
	}

	if !verif(data, &w1.PublicKey, sig2) {
		t.Fatal("Sig2 can't be checked with w1")
	}
}

func TestSaveAndLoadWrongWallet(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-*.wallet")
	if err != nil {
		panic(err)
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name())

	w1, err := wallet.New(tmpfile.Name())
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	w2, err := wallet.New(tmpfile.Name() + "nope")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name() + "nope")

	// w1 should be the same as w2
	data := []byte("testing data")

	sig1, err := sign(data, w1)
	if err != nil {
		panic(err)
	}

	if verif(data, &w2.PublicKey, sig1) {
		t.Fatal("Sig1 can be checked with diffrent w2")
	}

	sig2, err := sign(data, w2)
	if err != nil {
		panic(err)
	}

	if verif(data, &w1.PublicKey, sig2) {
		t.Fatal("Sig2 can be checked with diffrent w1")
	}
}
