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
		LocalAddr:   ":" + os.Getenv("VAL_PORT"),
		RemoteAddrs: os.Args[1:],
	}, privkey, currency, func(in, out chan core.Event) (pin, pout chan core.Event) {
		pin = make(chan core.Event, 256)
		pout = make(chan core.Event, 256)
		go func() {
			for i := range pin {
				log.Printf("EVENT pin -> in: %T", i)
				in <- i
			}
			close(in)
		}()
		go func() {
			for o := range out {
				log.Printf("EVENT out -> pout: %T", o)
				pout <- o
			}
			close(pout)
		}()
		return
	})

	log.Print("configured to run on port: " + os.Getenv("VAL_PORT"))

	backend.Start()
	defer backend.Stop()

	endpoint.Start(":" + os.Getenv("EP_PORT"))
}
