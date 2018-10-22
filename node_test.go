package gossipnet_test

import (
	"encoding/json"
	"github.com/Lywel/go-gossipnet"
	"sync"
	"testing"
)

type pbftMsg struct {
	Type  int
	Msg   string
	Other string
}

type idMsg struct {
	PublicKey string
}

// Message types
const (
	ibftMsg = iota
	joinMsg
	leaveMsg
)

func newJoinMsg(publicKey string) (*gossipnet.Message, error) {
	id := idMsg{publicKey}
	idBytes, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}

	join := gossipnet.Message{
		Data: idBytes,
		Type: joinMsg,
	}

	return &join, nil
}

func TestNew(t *testing.T) {
	var local, remote *gossipnet.Node
	var err error
	if local, err = gossipnet.New("pub1", "localhost:3000", []string{}); err != nil {
		t.Fatal(err)
	}
	if remote, err = gossipnet.New("pub2", "localhost:3001", []string{":3000"}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if err = local.Start(); err != nil {
		t.Fatal(err)
	}
	defer local.Stop()

	go func() {
		defer wg.Done()
		if err = remote.Start(); err != nil {
			t.Fatal(err)
		}
		defer remote.Stop()
	}()

	wg.Wait()
}

func TestBroadcast(t *testing.T) {
	var local, remote *gossipnet.Node
	var err error
	if local, err = gossipnet.New("pub1", "localhost:3000", []string{}); err != nil {
		t.Fatal(err)
	}
	if remote, err = gossipnet.New("pub2", "localhost:3001", []string{":3000"}); err != nil {
		t.Fatal(err)
	}

	if err = local.Start(); err != nil {
		t.Fatal(err)
	}
	defer local.Stop()

	ibft := pbftMsg{42, "TestNew msg", "TestNew other"}
	ibftBytes, err := json.Marshal(ibft)

	go func() {
		if err = remote.Start(); err != nil {
			t.Fatal(err)
		}
		defer remote.Stop()
		if err = remote.Broadcast(&gossipnet.Message{ibftBytes, 0}); err != nil {
			t.Fatal(err)
		}
	}()

	netMsg := <-local.OutChan
	var msg pbftMsg
	json.Unmarshal(netMsg.Data, &msg)
	if ibft != msg {
		t.Fatal("received message is diffrent :/")
	}
}
