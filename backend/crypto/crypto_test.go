package crypto_test

import (
	"."
	"crypto/ecdsa"
	"crypto/rand"
	eth "github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	var privkey *ecdsa.PrivateKey
	var err error
	if privkey, err = ecdsa.GenerateKey(eth.S256(), rand.Reader); err != nil {
		t.Fatal(err)
	}
	data := []byte("eth message")
	sig, err := crypto.Sign(data, privkey)
	if err := crypto.CheckSignature(data, sig, &privkey.PublicKey); err != nil {
		t.Fatal(err)
	}
}
