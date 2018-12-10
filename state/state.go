package state

import (
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
	"bitbucket.org/ventureslash/go-slash-currency/types"
)

type StateDB struct {
	stateObjects map[ibft.Address]StateObject
}

func New() *StateDB {
	return &StateDB{
		stateObjects: make(map[ibft.Address]StateObject),
	}
}

// ProcessBlock returns receitps of a block and update state
func (s *StateDB) ProcessBlock(b *types.Block) ([]*types.Receipt, error) {
	receipts := []*types.Receipt{}
	for _, t := range b.Transactions {
		res := uint64(1)
		sender := s.GetStateObject(t.From)
		receiver := s.GetStateObject(t.To)
		amount := t.Amount
		if !sender.SubBalance(amount) {
			res = 0
		} else {
			receiver.AddBalance(amount)
		}
		receipts = append(receipts, types.NewReceipt(t.Hash(), res))
	}
	n := b.Number().Uint64()
	if n != 0 && n%4320 == 0 { // 4320 = one week
		s.applyDemurrage()
	}
	return receipts, nil
}

func (s *StateDB) applyDemurrage() {
	for _, o := range s.GetStateObjects() {
		dem := new(big.Int).Div(o.GetBalance(), big.NewInt(3000))
		o.SubBalance(dem)
	}
}

// GetStateObject returns the state object associated to an address
func (s *StateDB) GetStateObject(addr ibft.Address) StateObject {
	state := s.stateObjects[addr]
	if state == nil {
		s.stateObjects[addr] = newStateObject()
		return s.stateObjects[addr]
	}
	return state
}

func (s *StateDB) GetStateObjects() map[ibft.Address]StateObject {
	return s.stateObjects
}

// GetBalance returns the balance associated to an address
func (s *StateDB) GetBalance(addr ibft.Address) *big.Int {
	return s.GetStateObject(addr).GetBalance()
}

type StateObject interface {
	GetBalance() *big.Int
	AddBalance(*big.Int) bool
	SubBalance(*big.Int) bool
	SetBalance(*big.Int)
}

type stateObject struct {
	balance *big.Int
}

func newStateObject() StateObject {
	return &stateObject{
		balance: big.NewInt(0),
	}
}

func (s *stateObject) GetBalance() *big.Int {
	return new(big.Int).Set(s.balance)
}

func (s *stateObject) AddBalance(amount *big.Int) bool {
	s.balance = new(big.Int).Add(s.balance, amount)
	return true
}

func (s *stateObject) SubBalance(amount *big.Int) bool {
	if s.balance.Cmp(amount) < 0 {
		return false
	}
	s.balance = new(big.Int).Sub(s.balance, amount)
	return true
}

func (s *stateObject) SetBalance(amount *big.Int) {
	s.balance = new(big.Int).Set(amount)
}
