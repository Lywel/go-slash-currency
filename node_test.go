package gossipnet_test

import (
	"bytes"
	"github.com/Lywel/go-gossipnet"
	"net"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	var local, remote *gossipnet.Node
	local = gossipnet.New("localhost:3000", []string{})
	remote = gossipnet.New("localhost:3001", []string{":3000"})

	var wg sync.WaitGroup
	wg.Add(1)

	if err := local.Start(); err != nil {
		t.Fatal(err)
	}
	defer local.Stop()

	go func() {
		defer wg.Done()
		if err := remote.Start(); err != nil {
			t.Fatal(err)
		}
		defer remote.Stop()
	}()

	wg.Wait()
}

type netMan struct {
	list map[net.Addr]bool
}

func (netMan) IsInteressted(net.Addr) bool {
	return true
}

func TestBroadcast(t *testing.T) {
	var local, remote *gossipnet.Node
	local = gossipnet.New("localhost:3000", []string{})
	remote = gossipnet.New("localhost:3001", []string{":3000"})

	if err := local.Start(); err != nil {
		t.Fatal(err)
	}
	defer local.Stop()

	ref := []byte("This is a test")

	go func() {
		if err := remote.Start(); err != nil {
			t.Fatal(err)
		}
		defer remote.Stop()

		remote.Broadcast(ref)
	}()

	received := <-local.EventChan()
	if bytes.Compare(received, ref) != 0 {
		t.Fatal("received msg expected to be '" + string(ref) + "' but got '" + string(received) + "' instead.")
	}
}
