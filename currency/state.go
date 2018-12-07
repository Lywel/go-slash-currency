package currency

import (
	"math/big"

	"bitbucket.org/ventureslash/go-ibft"
)

type StateDB struct {
	stateObjects map[ibft.Address]StateObject
}

func NewStateDB() *StateDB {
	return &StateDB{
		stateObjects: make(map[ibft.Address]StateObject),
	}
}

// GetStateObject returns the state object associated to an address
func (s StateDB) GetStateObject(addr ibft.Address) StateObject {
	state := s.stateObjects[addr]
	if state == nil {
		s.stateObjects[addr] = newStateObject()
		return s.stateObjects[addr]
	}
	return state
}

// GetBalance returns the balance associated to an address
func (s StateDB) GetBalance(addr ibft.Address) *big.Int {
	return s.GetStateObject(addr).GetBalance()
}

type StateObject interface {
	GetBalance() *big.Int
	AddBalance(*big.Int)
	SubBalance(*big.Int)
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

func (s *stateObject) AddBalance(amount *big.Int) {
	s.balance = new(big.Int).Add(s.balance, amount)
}

func (s *stateObject) SubBalance(amount *big.Int) {
	s.balance = new(big.Int).Sub(s.balance, amount)
}

func (s *stateObject) SetBalance(amount *big.Int) {
	s.balance = new(big.Int).Set(amount)
}
