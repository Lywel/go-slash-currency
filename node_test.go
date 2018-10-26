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
	localAddr := "127.0.0.1:3000"
	local = gossipnet.New(localAddr, []string{})
	remote = gossipnet.New("127.0.0.1:3001", []string{":3000"})

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

	event := <-local.EventChan()
	switch event.(type) {
	case gossipnet.DataEvent:
		dataEvent := event.(gossipnet.DataEvent)
		if bytes.Compare(dataEvent.Data, ref) != 0 {
			t.Fatal("received msg expected to be '" + string(ref) + "' but got '" + string(dataEvent.Data) + "' instead.")
		}
	case gossipnet.ListenEvent:
		listenEvent := event.(gossipnet.ListenEvent)
		if listenEvent.Addr != localAddr {
			t.Fatal("listening addr expected to be '" + string(localAddr) + "' but got '" + string(listenEvent.Addr) + "' instead.")
		}
	default:
	}
}
