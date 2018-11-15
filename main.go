package main

import (
	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-slash-currency/currency"
	"bitbucket.org/ventureslash/go-slash-currency/endpoint"
	"crypto/ecdsa"
	"crypto/rand"
	eth "github.com/ethereum/go-ethereum/crypto"
	"log"
	"os"
)

func main() {
	privkey, err := ecdsa.GenerateKey(eth.S256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	currency := &currency.Currency{}
	endpoint := endpoint.New()
	backend := backend.New(&backend.Config{
		LocalAddr:   ":" + os.Getenv("VAL_PORT"),
		RemoteAddrs: os.Args[1:],
	}, privkey, currency, endpoint.EventProxy())

	log.Print("configured to run on port: " + os.Getenv("VAL_PORT"))

	backend.Start()
	defer backend.Stop()

	endpoint.Start(":" + os.Getenv("EP_PORT"))
}
