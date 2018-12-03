package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"

	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-slash-currency/currency"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	eth "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/logger"
)

type addrSlice []string

func (as *addrSlice) String() string {
	return fmt.Sprint(*as)
}

func (i *addrSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var valAddrs addrSlice
	var syncAddrs addrSlice

	flag.Var(&valAddrs, "v", "address of a validator")
	flag.Var(&syncAddrs, "s", "address of a state provider")
	flag.Parse()

	logger.SetFlags(log.Lshortfile | log.Lmicroseconds)

	privkey, err := ecdsa.GenerateKey(eth.S256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	config := &backend.Config{
		LocalAddr:   ":" + os.Getenv("VAL_PORT"),
		RemoteAddrs: valAddrs,
	}

	currency := currency.New([]*types.Block{}, []*types.Transaction{}, config, privkey)

	log.Print("new currency created")
	remote := ""
	if len(syncAddrs) > 0 {
		remote = syncAddrs[0]
	}

	currency.SyncAndStart(remote)

}
