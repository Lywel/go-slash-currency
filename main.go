package main

import (
	"bitbucket.org/ventureslash/go-slash-currency/backend"
	"crypto/ecdsa"
	"crypto/rand"
	eth "github.com/ethereum/go-ethereum/crypto"
	"log"
	"os"
	"time"
)

func main() {
	privkey, err := ecdsa.GenerateKey(eth.S256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	backend := backend.New(&backend.Config{
		LocalAddr:   "0.0.0.0:8081",
		RemoteAddrs: os.Args[1:],
	}, privkey)

	backend.Start()
	defer backend.Stop()
	time.Sleep(240 * time.Second)
}