package main

import (
	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-ibft/core"
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
		LocalAddr:   ":" + os.Getenv("PORT"),
		RemoteAddrs: os.Args[1:],
	}, privkey, currency, func(in, out chan core.Event) (pin, pout chan core.Event) {
		pin, pout = in, out
		return
	})

	log.Print("configured to run on port: " + os.Getenv("PORT"))

	backend.Start()
	defer backend.Stop()

	endpoint.Start(":3000")
}
