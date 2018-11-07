package crypto

import (
	"crypto/ecdsa"

	"bitbucket.org/ventureslash/go-ibft/consensus"
	"github.com/ethereum/go-ethereum/crypto"
)

// Sign returns the signature of data from from privateKey
func Sign(data []byte, privkey *ecdsa.PrivateKey) ([]byte, error) {
	hashData := crypto.Keccak256([]byte(data))
	return crypto.Sign(hashData, privkey)
}

// PubkeyToAddress returns an Address from a ecdsa.PublicKey
func PubkeyToAddress(p ecdsa.PublicKey) consensus.Address {
	ethAddress := crypto.PubkeyToAddress(p)
	var a consensus.Address
	copy(a[consensus.AddressLength-len(ethAddress):], ethAddress[:])
	return a
}
