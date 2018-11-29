package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"log"
	"os"

	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-slash-currency/currency"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	eth "github.com/ethereum/go-ethereum/crypto"
)

func main() {
	privkey, err := ecdsa.GenerateKey(eth.S256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	config := &backend.Config{
		LocalAddr:   ":" + os.Getenv("VAL_PORT"),
		RemoteAddrs: os.Args[1:],
	}

	currency := currency.New([]*types.Block{}, []*types.Transaction{}, config, privkey)
	currency.Start()

}
