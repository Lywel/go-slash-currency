package main

import (
	"bitbucket.org/ventureslash/go-ibft/backend"
	"bitbucket.org/ventureslash/go-slash-currency/currency"
	"bitbucket.org/ventureslash/go-slash-currency/types"
	"bitbucket.org/ventureslash/go-slash-currency/wallet"
	"flag"
	"fmt"
	"github.com/google/logger"
	"log"
	"os"
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

	var walletPath = flag.String("w", "./slash-currency.wallet", "wallet file path")
	flag.Var(&valAddrs, "v", "address of a validator")
	flag.Var(&syncAddrs, "s", "address of a state provider")
	flag.Parse()

	logger.SetFlags(log.Lshortfile | log.Lmicroseconds)

	wallet, err := wallet.New(*walletPath)
	if err != nil {
		panic(err)
	}

	config := &backend.Config{
		LocalAddr:   ":" + os.Getenv("VAL_PORT"),
		RemoteAddrs: valAddrs,
	}

	currency := currency.New([]*types.Block{}, []*types.Transaction{}, config, wallet)

	remote := ""
	if len(syncAddrs) > 0 {
		remote = syncAddrs[0]
	}

	currency.SyncAndStart(remote)

}
