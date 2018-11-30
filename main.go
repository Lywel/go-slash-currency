package main

import (
	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-slash-currency/currency"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"crypto/ecdsa"
	"crypto/rand"
	"flag"
	eth "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/logger"
	"log"
	"os"
)

func main() {
	flag.Parse()
	logger.SetFlags(log.Lshortfile | log.Lmicroseconds)
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
