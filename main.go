package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"log"
	"os"
	"time"

	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-slash-currency/currency"

	eth "github.com/ethereum/go-ethereum/crypto"
)

func main() {
	privkey, err := ecdsa.GenerateKey(eth.S256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	currency := &currency.Currency{}
	backend := backend.New(&backend.Config{
		LocalAddr:   ":" + os.Getenv("PORT"),
		RemoteAddrs: os.Args[1:],
	}, privkey, currency)

	log.Print("configured to run on port: " + os.Getenv("PORT"))

	backend.Start()
	defer backend.Stop()
	time.Sleep(240 * time.Second)
}
